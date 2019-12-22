package repository

import (
	"fmt"

	"github.com/isutton/orchid/pkg/orchid/orm"
)

// Nested is responsible by given a Unstructured object it's able to return any field path nested
// in object's data.
type Nested struct {
	schema *orm.Schema
	obj    map[string]interface{}
	lines  []orm.Entry
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
func (n *Nested) Extract(fieldPath []string) ([]orm.Entry, error) {
	decomposed := decomposePaths(n.schema, fieldPath)
	decomposedLen := len(decomposed)
	if decomposedLen == 0 {
		return nil, fmt.Errorf("empty field-path informed, data is not nested!")
	}

	n.lines = []orm.Entry{}
	if err := n.nestedExtract(n.obj, decomposed, []string{}); err != nil {
		return nil, err
	}
	return n.lines, nil
}

// NewNested instantiate a new Nested.
func NewNested(schema *orm.Schema, obj map[string]interface{}) *Nested {
	return &Nested{schema: schema, obj: obj}
}
