package orm

import (
	"fmt"
	"strings"
)

// SQL represent the SQL statements.
type SQL struct {
	schema *Schema // ORM schema
}

// valuesPlaceholders creates dollar based notation for the amount specified.
func (s *SQL) valuesPlaceholders(amount int) []string {
	placeholders := []string{}
	for i := 1; i <= amount; i++ {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
	}
	return placeholders
}

// Insert generates a slice of inserts following the same tables sequence. Inserts carry "returning"
// therefore should always return "id" column value.
func (s *SQL) Insert() []string {
	inserts := []string{}
	for _, table := range s.schema.Tables {
		columnNames := table.ColumNames()
		insert := fmt.Sprintf(
			"insert into %s (%s) values (%s) returning id",
			table.Name,
			strings.Join(columnNames, ", "),
			strings.Join(s.valuesPlaceholders(len(columnNames)), ", "),
		)
		inserts = append(inserts, insert)
	}
	return inserts
}

// Select generates a select statement for the schema.
func (s *SQL) Select() string {
	columns := []string{}
	from := []string{}
	where := []string{}

	tables := s.schema.TablesReversed()
	for _, table := range tables {
		tableName := table.Name
		for _, column := range table.ColumNames() {
			columns = append(columns, fmt.Sprintf("%s.%s", tableName, column))
		}
		from = append(from, tableName)
		for _, constraint := range table.Constraints {
			if constraint.Type != PgConstraintFK {
				continue
			}
			where = append(where, fmt.Sprintf(
				"%s.%s=%s.%s",
				tableName,
				constraint.ColumnName,
				constraint.RelatedTableName,
				constraint.RelatedColumnName,
			))
		}
	}

	return fmt.Sprintf("select %s from %s where %s",
		strings.Join(columns, ", "), strings.Join(from, ", "), strings.Join(where, " AND "))
}

// CreateTables return the statements needed to create Schema tables, leaving primary Schema table
// as last.
func (s *SQL) CreateTables() []string {
	statements := []string{}
	for _, table := range s.schema.Tables {
		statements = append(statements, table.String())
	}
	return statements
}

// NewSQL instantiate an SQL.
func NewSQL(schema *Schema) *SQL {
	return &SQL{schema: schema}
}
