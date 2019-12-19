package repository

import (
	"fmt"

	"github.com/isutton/orchid/pkg/orchid/orm"
)

type Nested struct {
	schema *orm.Schema
	obj    map[string]interface{}
	lines  []map[string]interface{}
}

// decomposePaths based on field path, it will check against schema for one-to-many relationships,
// in this case the path is decomposed to reflect nested data. It returns a slice of slices, on
// which inner level represents individual paths.
func (n *Nested) decomposePaths(fieldPath []string) [][]string {
	decomposed := [][]string{}
	currentPath := []string{}
	for _, field := range fieldPath {
		currentPath = append(currentPath, field)

		if n.schema.HasOneToMany(currentPath) {
			decomposed = append(decomposed, currentPath)
			currentPath = []string{}
		}
	}
	if len(currentPath) > 0 {
		decomposed = append(decomposed, currentPath)
	}
	return decomposed
}

// nestedExtract recursively extract decomposed-paths in order to return a list of lines, in other
// words a list of map[string]interface{}. It can return error on reading unstructured data.
func (n *Nested) nestedExtract(
	obj map[string]interface{},
	decomposedPaths [][]string,
	startingAt []string,
) error {
	for position, decomposed := range decomposedPaths {
		nextPosition := position + 1
		startingAt = append(startingAt, decomposed...)

		// when dealing with decomposed paths the first entry will start from the root, while the
		// upcoming entries will be based in a different root
		var fieldPath []string
		if len(startingAt) > 0 {
			fieldPath = decomposed
		} else {
			fieldPath = startingAt
		}

		// one-to-many is represented as an slice of maps, therefore it will trigger recursion
		// on each item found
		if n.schema.HasOneToMany(startingAt) && !n.schema.IsKV(startingAt) {
			slice, err := nestedSlice(obj, fieldPath)
			if err != nil {
				return err
			}

			for _, sliceItem := range slice {
				item := sliceItem.(map[string]interface{})

				if nextPosition < len(decomposedPaths) {
					err = n.nestedExtract(item, decomposedPaths[nextPosition:], startingAt)
					if err != nil {
						return err
					}
				} else {
					n.lines = append(n.lines, item)
				}
			}

			// stoping an next recursive call after the first one-to-many flag.
			break
		}

		// dealing with the left-over objects
		item, err := nestedMap(obj, startingAt)
		if err != nil {
			return err
		}
		n.lines = append(n.lines, item)
	}
	return nil
}

// Extract recursively extract field-path from unstructured object, returning an array of maps,
// representing the lines found for that entity. It can return errors on navigating unstructured.
func (n *Nested) Extract(fieldPath []string) ([]map[string]interface{}, error) {
	decomposed := n.decomposePaths(fieldPath)
	decomposedLen := len(decomposed)
	if decomposedLen == 0 {
		return nil, fmt.Errorf("empty field-path informed, data is not nested!")
	}

	n.lines = make([]map[string]interface{}, 0)
	if err := n.nestedExtract(n.obj, decomposed, []string{}); err != nil {
		return nil, err
	}
	return n.lines, nil
}

func NewNested(schema *orm.Schema, obj map[string]interface{}) *Nested {
	return &Nested{schema: schema, obj: obj}
}
