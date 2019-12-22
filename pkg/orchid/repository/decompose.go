package repository

import (
	"github.com/isutton/orchid/pkg/orchid/orm"
)

// decomposePaths based on field path, it will check against schema for one-to-many relationships,
// in this case the path is decomposed to reflect nested data. It returns a slice of slices, on
// which inner level represents individual paths.
func decomposePaths(s *orm.Schema, fieldPath []string) [][]string {
	decomposed := [][]string{}
	currentPath := []string{}
	for _, field := range fieldPath {
		currentPath = append(currentPath, field)

		if s.HasOneToMany(currentPath) {
			decomposed = append(decomposed, currentPath)
			currentPath = []string{}
		}
	}
	if len(currentPath) > 0 {
		decomposed = append(decomposed, currentPath)
	}
	return decomposed
}
