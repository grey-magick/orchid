package orm

import (
	"fmt"
)

// Column represents columns.
type Column struct {
	Name string // column name
	Type string // column type
}

// String print out column and type.
func (c *Column) String() string {
	return fmt.Sprintf("%s %s", c.Name, c.Type)
}
