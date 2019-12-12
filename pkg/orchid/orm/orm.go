package orm

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/isutton/orchid/pkg/orchid/runtime"
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

func (o *ORM) Create(schema *Schema, obj runtime.Object) error {
	// _ := NewSQL(schema)
	//
	// // transaction begin
	// for _, table := range schema.Tables {
	// 	for _, _ := range table.Columns {
	// 	}
	// }
	// // transaction commit

	return nil
}

// NewORM instantiate an ORM.
func NewORM(connStr string) *ORM {
	return &ORM{connStr: connStr}
}
