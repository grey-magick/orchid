package orm

import (
	"fmt"
	"strings"
)

// Table represents database table with columns and constraints.
type Table struct {
	Name        string        // table name
	Columns     []*Column     // table columns
	Constraints []*Constraint // constraints
}

// GetColumn return the instance of column based on name.
func (t *Table) GetColumn(name string) *Column {
	for _, column := range t.Columns {
		if name == column.Name {
			return column
		}
	}
	return nil
}

// AddConstraint add a new constraint.
func (t *Table) AddConstraint(constraint *Constraint) {
	t.Constraints = append(t.Constraints, constraint)
}

// AddSerialPK add a new column as primary-key, using serial8 type.
func (t *Table) AddSerialPK() {
	columnName := "id"
	t.AddColumn(&Column{Name: columnName, Type: PgTypeSerial8})
	t.AddConstraint(&Constraint{Type: PgConstraintPK, ColumnName: columnName})
}

// AddBigIntPK add a new column as a primary-key, using BigInt type.
func (t *Table) AddBigIntPK() {
	columnName := "id"
	t.AddColumn(&Column{Name: columnName, Type: PgTypeBigInt})
	t.AddConstraint(&Constraint{Type: PgConstraintPK, ColumnName: columnName})
}

// AddForeignKey adds a new column with foreign-key constraint.
func (t *Table) AddBigIntFK(columnName, relatedTableName string, notNull bool) {
	t.AddColumn(&Column{Name: columnName, Type: PgTypeBigInt, NotNull: notNull})
	t.AddConstraint(&Constraint{
		Type:              PgConstraintFK,
		ColumnName:        columnName,
		RelatedTableName:  relatedTableName,
		RelatedColumnName: "id",
	})
}

// AddColumn append a new column.
func (t *Table) AddColumn(column *Column) {
	t.Columns = append(t.Columns, column)
}

// ColumNames return a slice of column names.
func (t *Table) ColumNames() []string {
	names := []string{}
	for _, column := range t.Columns {
		names = append(names, column.Name)
	}
	return names
}

// String print out table creation SQL statement.
func (t *Table) String() string {
	columns := []string{}
	for _, column := range t.Columns {
		columns = append(columns, column.String())
	}
	constrains := []string{}
	for _, constraint := range t.Constraints {
		constrains = append(constrains, constraint.String())
	}
	return fmt.Sprintf("create table if not exists %s (%s, %s)",
		t.Name, strings.Join(columns, ", "), strings.Join(constrains, ", "))
}

// NewTable instantiate a new Table.
func NewTable(name string) *Table {
	return &Table{Name: name}
}
