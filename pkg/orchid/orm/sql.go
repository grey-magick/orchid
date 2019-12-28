package orm

import (
	"fmt"
	"strings"
)

// CreateDatabaseStatement returns create database statement with informed database.
func CreateDatabaseStatement(database string) string {
	return fmt.Sprintf("create database %s template 'template1'", database)
}

// SelectDatabaseStatement returns select statement to check if database exists.
func SelectDatabaseStatement() string {
	return "select 1 from pg_database where datname = $1"
}

// CreateSchemaStatement returns create schema statement, with informed search-path.
func CreateSchemaStatement(searchPath string) string {
	return fmt.Sprintf("create schema if not exists %s", searchPath)
}

// valuesPlaceholders creates dollar based notation for the amount specified.
func valuesPlaceholders(amount int) []string {
	placeholders := []string{}
	for i := 1; i <= amount; i++ {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
	}
	return placeholders
}

// InsertStatement generates a slice of inserts following the same tables sequence. Inserts carry
// "returning" therefore should always return "id" column value.
func InsertStatement(schema *Schema) []string {
	inserts := []string{}
	for _, table := range schema.Tables {
		columnNames := []string{}
		for _, column := range table.ColumNames() {
			columnNames = append(columnNames, fmt.Sprintf("\"%s\"", column))
		}

		insert := fmt.Sprintf(
			"insert into %s (%s) values (%s) returning %s",
			table.Name,
			strings.Join(columnNames, ", "),
			strings.Join(valuesPlaceholders(len(columnNames)), ", "),
			PKColumnName,
		)
		inserts = append(inserts, insert)
	}
	return inserts
}

// SelectStatement generates a select statement for the schema, using where clause informed. Where
func SelectStatement(schema *Schema, where []string) string {
	statementColumns := []string{}
	statementFrom := []string{}
	statementWhere := []string{"1=1"}

	for _, table := range schema.TablesReversed() {
		columns := []string{PKColumnName}
		columns = append(columns, table.ColumNames()...)

		for _, column := range columns {
			statementColumns = append(statementColumns,
				fmt.Sprintf("%s.\"%s\" as \"%s.%s\"", table.Hint, column, table.Hint, column))
		}

		for _, constraint := range table.Constraints {
			if constraint.Type != PgConstraintFK {
				continue
			}
			statementWhere = append(statementWhere, fmt.Sprintf("%s.\"%s\"=%s.\"%s\"",
				schema.GetHint(constraint.RelatedTableName),
				constraint.RelatedColumnName,
				table.Hint,
				constraint.ColumnName,
			))
		}

		statementFrom = append(statementFrom, fmt.Sprintf("%s %s", table.Name, table.Hint))
	}
	for i, clause := range where {
		statementWhere = append(statementWhere, fmt.Sprintf("%s=$%d", clause, i+1))
	}
	return fmt.Sprintf(
		"select %s from %s where %s",
		strings.Join(statementColumns, ", "),
		strings.Join(statementFrom, ", "),
		strings.Join(statementWhere, " and "),
	)
}

// CreateTablesStatement return the statements needed to create table and add foreign keys.
func CreateTablesStatement(schema *Schema) []string {
	createTables := []string{}
	for _, table := range schema.Tables {
		createTables = append(createTables, table.String())
	}
	return createTables
}
