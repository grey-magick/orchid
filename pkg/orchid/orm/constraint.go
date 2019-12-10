package orm

import (
	"fmt"
)

// Constraint represents constraints.
type Constraint struct {
	Type string // constraint type
	Expr string // constraint expression
}

// String print out constraint and expression.
func (c *Constraint) String() string {
	return fmt.Sprintf("%s %s", c.Type, c.Expr)
}
