package orm

import (
	"fmt"
	"strings"

	sqlfmt "github.com/otaviof/go-sqlfmt/pkg/sqlfmt"
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

// hintedColumns returns a slice of column names using table hint. Does not include foreign-keys.
func hintedColumns(table *Table) []string {
	columnNames := []string{PKColumnName}
	columnNames = append(columnNames, table.ColumNames()...)
	columns := []string{}
	for _, column := range columnNames {
		columns = append(columns,
			fmt.Sprintf("%s.\"%s\" as \"%s.%s\"", table.Hint, column, table.Hint, column))
	}
	return columns
}

// leftJoin creates a left-join clause using different approaches depending on the one-to-many table
// flag, when one-to-many the current table is included as the main left join table, when one-to-one
// the related table is the primary table
func leftJoin(schema *Schema, table *Table, constraint *Constraint, related *Table) string {
	if schema.HasOneToMany(table.Path) {
		return fmt.Sprintf(
			"left join %s %s on %s.%s=%s.%s",
			table.Name, table.Hint,
			table.Hint, constraint.ColumnName,
			related.Hint, constraint.RelatedColumnName,
		)
	}
	return fmt.Sprintf(
		"left join %s %s on %s.%s=%s.%s",
		related.Name, related.Hint,
		table.Hint, constraint.ColumnName,
		related.Hint, constraint.RelatedColumnName,
	)
}

// SelectStatement generates a select statement based on schema, using the primary schema table
// as from, and other tables as left-join entries. It can return error when tables are not found.
func SelectStatement(schema *Schema, where []string) (string, error) {
	mainTable, err := schema.GetTable(schema.Name)
	if err != nil {
		return "", err
	}
	// preparing statement "from" clause based on main schema table
	from := []string{fmt.Sprintf("%s %s", mainTable.Name, mainTable.Hint)}

	leftJoins := []string{}
	columns := []string{}
	for _, table := range schema.Tables {
		columns = append(columns, hintedColumns(table)...)

		for _, constraint := range table.Constraints {
			if constraint.Type != PgConstraintFK {
				continue
			}

			related, err := schema.GetTable(constraint.RelatedTableName)
			if err != nil {
				return "", err
			}

			// in order to keep the correct sql statement sequence, when dealing with a table having
			// one-to-many relationships, moving the statement to the end of left-joins group, while
			// when not having one-to-many, it will be prepended
			leftJoinStmt := leftJoin(schema, table, constraint, related)
			if schema.HasOneToMany(table.Path) {
				leftJoins = append(leftJoins, leftJoinStmt)
			} else {
				leftJoins = StringSlicePrepend(leftJoins, leftJoinStmt)
			}
		}
	}

	statement := fmt.Sprintf(
		"select %s from %s",
		strings.Join(columns, ", "),
		strings.Join(from, ", "),
	)
	if len(leftJoins) > 0 {
		statement = fmt.Sprintf("%s %s", statement, strings.Join(leftJoins, " "))
	}
	if len(where) > 0 {
		statement = fmt.Sprintf("%s where %s", statement, strings.Join(where, " and "))
	}
	return statement, nil
}

func FormatStatement(statement string) string {
	opts := &sqlfmt.Options{Distance: 0}
	formatted, _ := sqlfmt.Format(statement, opts)
	return formatted
}

// CreateTablesStatement return the statements needed to create table and add foreign keys.
func CreateTablesStatement(schema *Schema) []string {
	createTables := []string{}
	for _, table := range schema.Tables {
		createTables = append(createTables, table.String())
	}
	return createTables
}
