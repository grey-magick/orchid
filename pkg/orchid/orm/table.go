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
	Path        []string      // path to the node in the Orchid object
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

// IsPrimaryKey inspect table's constraints to check if informed column name is primary key.
func (t *Table) IsPrimaryKey(columnName string) bool {
	for _, constraint := range t.Constraints {
		if constraint.ColumnName == columnName && constraint.Type == PgConstraintPK {
			return true
		}
	}
	return false
}

func (t *Table) IsForeignKey(columnName string) bool {
	for _, constraint := range t.Constraints {
		if constraint.ColumnName == columnName && constraint.Type == PgConstraintFK {
			return true
		}
	}
	return false
}

// ForeignKeyTable in case of columnName being a foreign key, returning the table name it points to.
func (t *Table) ForeignKeyTable(columnName string) string {
	for _, constraint := range t.Constraints {
		if constraint.ColumnName == columnName && constraint.Type == PgConstraintFK {
			return constraint.RelatedTableName
		}
	}
	return ""
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
func (t *Table) AddBigIntFK(
	columnName string,
	relatedTableName string,
	relatedColumnName string,
	notNull bool,
) {
	t.AddColumn(&Column{
		Name:         columnName,
		Type:         PgTypeBigInt,
		OriginalType: JSTypeObject,
		NotNull:      notNull,
	})
	t.AddConstraint(&Constraint{
		Type:              PgConstraintFK,
		ColumnName:        columnName,
		RelatedTableName:  relatedTableName,
		RelatedColumnName: relatedColumnName,
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
		if column.Type == PgTypeSerial8 {
			continue
		}
		names = append(names, column.Name)
	}
	return names
}

// String print out table creation SQL statement, and alter-table statement to include foreign keys
// later in the process.
func (t *Table) String() (string, string) {
	columns := []string{}
	for _, column := range t.Columns {
		columns = append(columns, column.String())
	}

	foreignKeys := []string{}
	constrains := []string{}
	for _, constraint := range t.Constraints {
		if constraint.Type == PgConstraintFK {
			foreignKeys = append(foreignKeys, constraint.String())
		} else {
			constrains = append(constrains, constraint.String())
		}
	}

	createTable := fmt.Sprintf("create table if not exists %s (%s, %s)",
		t.Name, strings.Join(columns, ", "), strings.Join(constrains, ", "))
	if len(foreignKeys) == 0 {
		return createTable, ""
	}

	alterTable := fmt.Sprintf("alter table %s add %s", t.Name, strings.Join(foreignKeys, ", add "))
	return createTable, alterTable

}

// NewTable instantiate a new Table.
func NewTable(name string, tablePath []string) *Table {
	return &Table{Name: name, Path: tablePath}
}
