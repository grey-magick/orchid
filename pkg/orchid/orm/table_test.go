package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable_New(t *testing.T) {
	t.Run("Hint", func(t *testing.T) {
		table := NewTable("test")
		assert.Equal(t, "t", table.Hint)

		table = NewTable("test_test_test")
		assert.Equal(t, "ttt", table.Hint)

		table = NewTable("TEST_TEST_TEST")
		assert.Equal(t, "ttt", table.Hint)
		assert.Equal(t, "test_test_test", table.Name)
	})

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

		assert.True(t, table.IsPrimaryKey(PKColumnName))
	})

	t.Run("AddBigIntFK", func(t *testing.T) {
		table := NewTable("test")
		table.AddBigIntFK("column", "onTable", PKColumnName, false)

		assert.Len(t, table.Columns, 1)
		assert.Equal(t, PgTypeBigInt, table.Columns[0].Type)

		assert.Len(t, table.Constraints, 1)
		assert.Equal(t, PgConstraintFK, table.Constraints[0].Type)

		assert.Equal(t, "onTable", table.ForeignKeyTable("column"))
	})

	t.Run("column-names...", func(t *testing.T) {
		table := NewTable("test")
		table.AddBigIntFK("column", "onTable", PKColumnName, true)

		assert.NotNil(t, table.GetColumn("column"))

		assert.Equal(t, []string{"column"}, table.ColumNames())
		assert.Equal(t, []string{}, table.ColumnNamesStripped())

		assert.Len(t, table.ForeignKeys(), 1)
		assert.Equal(t, "onTable", table.ForeignKeyTable("column"))
	})

	t.Run("String", func(t *testing.T) {
		table := NewTable("test")
		createTable := table.String()

		assert.NotEmpty(t, createTable)
		assert.Contains(t, createTable, "create table")
	})
}
