package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable_New(t *testing.T) {
	t.Run("AddSerialPK", func(t *testing.T) {
		table := NewTable("test")
		table.AddSerialPK()

		assert.Len(t, table.Columns, 1)
		assert.Equal(t, PgTypeSerial8, table.Columns[0].Type)

		assert.Len(t, table.Constraints, 1)
		assert.Equal(t, PgConstraintPK, table.Constraints[0].Type)
	})

	t.Run("AddBigIntPK", func(t *testing.T) {
		table := NewTable("test")
		table.AddBigIntPK()

		assert.Len(t, table.Columns, 1)
		assert.Equal(t, PgTypeBigInt, table.Columns[0].Type)

		assert.Len(t, table.Constraints, 1)
		assert.Equal(t, PgConstraintPK, table.Constraints[0].Type)
	})

	t.Run("AddBigIntFK", func(t *testing.T) {
		table := NewTable("test")
		table.AddBigIntFK("column", "onTable", false)

		assert.Len(t, table.Columns, 1)
		assert.Equal(t, PgTypeBigInt, table.Columns[0].Type)

		assert.Len(t, table.Constraints, 1)
		assert.Equal(t, PgConstraintFK, table.Constraints[0].Type)
	})

	t.Run("ColumnNames", func(t *testing.T) {
		table := NewTable("test")
		table.AddBigIntFK("column", "onTable", true)

		assert.Equal(t, []string{"column"}, table.ColumNames())
	})

	t.Run("String", func(t *testing.T) {
		table := NewTable("test")

		assert.NotEmpty(t, table.String())
	})
}
