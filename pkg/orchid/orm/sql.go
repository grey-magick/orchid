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
		if s.schema.Name != table.Name {
			statements = append(statements, table.String())
		}
	}
	return append(statements, s.schema.Tables[s.schema.Name].String())
}

// NewSQL instantiate an SQL.
func NewSQL(schema *Schema) *SQL {
	return &SQL{schema: schema}
}
