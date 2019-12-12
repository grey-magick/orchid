package orm

import (
	"fmt"
)

// Constraint represents constraints.
type Constraint struct {
	Type              string // constraint type
	ColumnName        string // local column name
	RelatedTableName  string // related table name
	RelatedColumnName string // related table's column name
}

// String print out constraint and expression.
func (c *Constraint) String() string {
	switch c.Type {
	case PgConstraintFK:
		return fmt.Sprintf("%s (%s) references %s (%s)",
			c.Type, c.ColumnName, c.RelatedTableName, c.RelatedColumnName)
	default:
		return fmt.Sprintf("%s (%s)", c.Type, c.ColumnName)
	}
}
