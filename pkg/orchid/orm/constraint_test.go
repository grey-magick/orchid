package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstraint(t *testing.T) {
	for _, pgType := range []string{PgConstraintFK, PgConstraintPK, PgConstraintUnique} {
		t.Run("ForeignKey", func(t *testing.T) {
			constraint := &Constraint{
				Type:              pgType,
				ColumnName:        "column",
				RelatedTableName:  "onTable",
				RelatedColumnName: "onColumn",
			}
			assert.NotEmpty(t, constraint.String)
		})
	}
}
