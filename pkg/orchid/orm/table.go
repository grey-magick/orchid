package orm

import (
	"fmt"
	"strings"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
)

// Table represents database table with columns and constraints.
type Table struct {
	Name        string        // table name
	Hint        string        // short table name
	Path        []string      // path to the node in the Orchid object
	Columns     []*Column     // table columns
	Constraints []*Constraint // constraints
	OneToMany   bool          // meant for ene-to-many relationship
	KV          bool          // meant for key-value store
}

// PKColumnName primary-key column name
const PKColumnName = "id"

// buildHint create a hint out of a name, by splitting on underscore and using its first character.
func (t *Table) buildHint() {
	var short string
	for _, section := range strings.Split(t.Name, "_") {
		short = fmt.Sprintf("%s%s", short, string(section[0]))
	}
	t.Hint = strings.ToLower(short)
}

// hasContraint check slice of contraint for a given constraint type and column name.
func (t *Table) hasContraint(contratintType, columnName string) bool {
	for _, constraint := range t.Constraints {
		if constraint.ColumnName == columnName && constraint.Type == contratintType {
			return true
		}
	}
	return false
}

// AddColumn append a new column.
func (t *Table) AddColumn(column *Column) {
	t.Columns = append(t.Columns, column)
}

// AddConstraint add a new constraint.
func (t *Table) AddConstraint(constraint *Constraint) {
	t.Constraints = append(t.Constraints, constraint)
}

// AddSerialPK add a new column as primary-key, using serial8 type.
func (t *Table) AddSerialPK() {
	t.AddColumn(&Column{Name: PKColumnName, Type: PgTypeSerial8})
	t.AddConstraint(&Constraint{Type: PgConstraintPK, ColumnName: PKColumnName})
}

// AddBigIntPK add a new column as a primary-key, using BigInt type.
func (t *Table) AddBigIntPK() {
	t.AddColumn(&Column{Name: PKColumnName, Type: PgTypeBigInt})
	t.AddConstraint(&Constraint{Type: PgConstraintPK, ColumnName: PKColumnName})
}

// AddForeignKey adds a new column with foreign-key constraint.
func (t *Table) AddBigIntFK(
	columnName string,
	relatedTableName string,
	relatedColumnName string,
	notNull bool,
) {
	t.AddColumn(&Column{Name: columnName, Type: PgTypeBigInt, JSType: jsc.Object, NotNull: notNull})
	t.AddConstraint(&Constraint{
		Type:              PgConstraintFK,
		ColumnName:        columnName,
		RelatedTableName:  relatedTableName,
		RelatedColumnName: relatedColumnName,
	})
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

// String print out table creation SQL statement, and alter-table statement to include foreign keys
// IsPrimaryKey inspect table's constraints to check if informed column name is primary key.
func (t *Table) IsPrimaryKey(columnName string) bool {
	return t.hasContraint(PgConstraintPK, columnName)
}

// IsForeignKey inspect constraints to check if a foreign-key is set to column name.
func (t *Table) IsForeignKey(columnName string) bool {
	return t.hasContraint(PgConstraintFK, columnName)
}

// ColumNames return a slice of column names, without primary key included.
func (t *Table) ColumNames() []string {
	names := make([]string, 0, len(t.Columns)-1)
	for _, column := range t.Columns {
		if column.Type == PgTypeSerial8 {
			continue
		}
		names = append(names, column.Name)
	}
	return names
}

// ColumnNamesStripped list of column names where primary-key, foregin-keys and other special
// column names are not listed.
func (t *Table) ColumnNamesStripped() []string {
	names := make([]string, 0, len(t.Columns)-1)
	for _, column := range t.ColumNames() {
		if t.IsForeignKey(column) {
			continue
		}
		names = append(names, column)
	}
	return names
}

// ForeignKeys return a list of foreign-keys constraints for table.
func (t *Table) ForeignKeys() []*Constraint {
	fks := []*Constraint{}
	for _, constraint := range t.Constraints {
		if constraint.Type != PgConstraintFK {
			continue
		}
		fks = append(fks, constraint)
	}
	return fks
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

// String returns the respective create table statement.
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
	table := &Table{Name: strings.ToLower(name)}
	table.buildHint()
	return table
}
