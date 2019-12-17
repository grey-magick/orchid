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
	leftJoin := []string{}

	tables := s.schema.TablesReversed()
	for _, table := range tables {
		tableName := table.Name
		for _, column := range table.ColumNames() {
			columns = append(columns, fmt.Sprintf("%s.%s", tableName, column))
		}
		if !table.OneToMany {
			from = append(from, tableName)
		}
		for _, constraint := range table.Constraints {
			if table.OneToMany && constraint.RelatedTableName != "" {
				leftJoin = append(leftJoin, fmt.Sprintf(
					"left join %s on %s.id=%s.%s",
					constraint.RelatedTableName,
					table.Name,
					constraint.RelatedTableName,
					constraint.RelatedColumnName,
				))
			} else {
				continue
				where = append(where, fmt.Sprintf(
					"%s.%s=%s.%s",
					tableName,
					constraint.ColumnName,
					constraint.RelatedTableName,
					constraint.RelatedColumnName,
				))
			}
		}
	}

	return fmt.Sprintf(
		"select %s\nfrom %s\n%s\nwhere %s\n",
		strings.Join(columns, ", "),
		strings.Join(from, ", "),
		strings.Join(leftJoin, " "),
		strings.Join(where, " AND "),
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
