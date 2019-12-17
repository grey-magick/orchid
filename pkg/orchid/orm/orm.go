package orm

import (
	"database/sql"
	"fmt"
	"log"

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
	for _, statement := range sqlLib.CreateTables() {
		log.Printf("statement='%s'", statement)
		_, err := o.db.Query(statement)
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

func (o *ORM) interpolate(
	table *Table,
	arguments []interface{},
	cachedIDs map[string]int64,
) ([]interface{}, error) {
	argumentWithFK := []interface{}{}
	pos := 0
	for _, column := range table.Columns {
		if table.IsPrimaryKey(column.Name) {
			continue
		}
		if targetFKTable := table.ForeignKeyTable(column.Name); targetFKTable != "" {
			foreingKeyID, found := cachedIDs[targetFKTable]
			if !found {
				return nil, fmt.Errorf("unable to find primary-key in cache '%#v'", cachedIDs)
			}
			argumentWithFK = append(argumentWithFK, foreingKeyID)
		} else {
			argumentWithFK = append(argumentWithFK, arguments[pos])
			pos += 1
		}
	}
	return argumentWithFK, nil
}

func (o *ORM) Create(schema *Schema, argumentsPerTable map[string][][]interface{}) error {
	sqlLib := NewSQL(schema)
	statements := sqlLib.Insert()
	cachedIDs := make(map[string]int64, len(statements))

	txn, err := o.db.Begin()
	if err != nil {
		return err
	}

	for i, table := range schema.Tables {
		statement := statements[i]
		argumentsSlice, found := argumentsPerTable[table.Name]
		if !found {
			continue
		}

		for _, arguments := range argumentsSlice {
			// in case the case of arguments for this table being less than expected, completing the
			// slice with foreign-key cached IDs
			if len(arguments) == 0 {
				log.Print("[WARN] arguments is empty!!")
				continue
			}

			if len(arguments) < len(table.Columns)-1 {
				if arguments, err = o.interpolate(table, arguments, cachedIDs); err != nil {
					return err
				}
			}

			log.Printf("statement='%s', arguments='%#v'", statement, arguments)

			var primaryKeyValue int64
			if err = txn.QueryRow(statement, arguments...).Scan(&primaryKeyValue); err != nil {
				return err
			}
			log.Printf("primary-key='%#v'", primaryKeyValue)
			cachedIDs[table.Name] = primaryKeyValue
		}
	}

	return txn.Commit()
}

// NewORM instantiate an ORM.
func NewORM(connStr string) *ORM {
	return &ORM{connStr: connStr}
}
