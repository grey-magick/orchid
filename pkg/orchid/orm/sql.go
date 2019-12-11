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

// Insert generates a map of inserts per table.
func (s *SQL) Insert() map[string]string {
	inserts := map[string]string{}
	for _, table := range s.schema.Tables {
		columnNames := table.ColumNames()
		insert := fmt.Sprintf(
			"insert into %s (%s) values (%s)",
			table.Name,
			strings.Join(columnNames, ", "),
			strings.Join(s.valuesPlaceholders(len(columnNames)), ", "),
		)
		inserts[table.Name] = insert
	}
	return inserts
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
