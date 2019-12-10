package orm

// SQL represent the SQL statements.
type SQL struct {
	schema *Schema // ORM schema
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
