package orm

import (
	"fmt"
	"strings"
)

// SQL represent the SQL statements.
type SQL struct {
	schema *Schema // ORM schema
}

// hint create a hint out of a name, by splitting on underscore and using the first charactere.
func (s *SQL) hint(name string) string {
	var short string
	for _, section := range strings.Split(name, "_") {
		short = fmt.Sprintf("%s%s", short, string(section[0]))
	}
	return strings.ToLower(short)
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

	for _, table := range s.schema.TablesReversed() {
		tableName := table.Name

		for _, column := range table.ColumNames() {
			columns = append(columns, fmt.Sprintf("%s.%s", s.hint(tableName), column))
		}

		for _, constraint := range table.Constraints {
			if constraint.Type != PgConstraintFK {
				continue
			}
			where = append(where, fmt.Sprintf("%s.%s=%s.%s",
				s.hint(constraint.RelatedTableName),
				constraint.RelatedColumnName,
				s.hint(tableName),
				constraint.ColumnName,
			))
		}

		from = append(from, fmt.Sprintf("%s %s", tableName, s.hint(tableName)))
	}

	return fmt.Sprintf(
		"select %s from %s where %s",
		strings.Join(columns, ", "),
		strings.Join(from, ", "),
		strings.Join(where, " and "),
	)
}

// CreateTables return the statements needed to create table and add foreign keys.
func (s *SQL) CreateTables() []string {
	createTables := []string{}
	for _, table := range s.schema.Tables {
		createTables = append(createTables, table.String())
	}
	return createTables
}

// NewSQL instantiate an SQL.
func NewSQL(schema *Schema) *SQL {
	return &SQL{schema: schema}
}
