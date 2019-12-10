package orm

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// ORM ...
type ORM struct {
	connStr string  // database adapter connection string
	db      *sql.DB // database adapter instance
}

// CreateSchemaTables create tables for a schema.
func (o *ORM) CreateSchemaTables(schema *Schema) error {
	sqlLib := NewSQL(schema)
	for _, stmt := range sqlLib.CreateTables() {
		_, err := o.db.Query(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// Connect with the database, instantiate the connection.
func (o *ORM) Connect() error {
	var err error
	o.db, err = sql.Open("postgres", o.connStr)
	return err
}

// NewORM instantiate an ORM.
func NewORM(connStr string) *ORM {
	return &ORM{connStr: connStr}
}
