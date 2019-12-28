package orm

import (
	"database/sql"
	"fmt"

	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
	"k8s.io/apimachinery/pkg/types"

	"github.com/isutton/orchid/pkg/orchid/config"
)

// ORM represents the data abastraction layer.
type ORM struct {
	logger     logr.Logger    // logger instance
	database   string         // database name
	searchPath string         // database schema name
	config     *config.Config // configuration instance
	DB         *sql.DB        // database adapter instance
}

// driverName database driver
const driverName = "postgres"

type List []interface{}
type MappedList map[string]List
type MappedMatrix map[string][]List

type Entry map[string]interface{}
type EntryMap map[string]Entry
type MappedEntries map[string][]Entry

// createDatabase create an PostgreSQL database.
func (o *ORM) createDatabase() error {
	var exists int = 0
	err := o.DB.QueryRow(SelectDatabaseStatement(), o.database).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if exists == 1 {
		o.logger.Info("Database already exists!")
		return nil
	}

	o.logger.Info("Creating database...")
	_, err = o.DB.Exec(CreateDatabaseStatement(o.database))
	return err
}

// createSchema create an PostgreSQL schema.
func (o *ORM) createSchema() error {
	_, err := o.DB.Exec(CreateSchemaStatement(o.searchPath))
	return err
}

// Bootstrap initial connection to make sure database is present, and a second connection to then
// create schema, making sure subsequent queries will use the schema as search-path.
func (o *ORM) Bootstrap() error {
	// connecting with a privileged user first to create database and schema
	if err := o.connect("postgres", "public"); err != nil {
		return err
	}

	if err := o.createDatabase(); err != nil {
		return err
	}

	// closing current connection in order to open a new one on specific database
	if err := o.DB.Close(); err != nil {
		return err
	}

	if err := o.connect(o.database, o.searchPath); err != nil {
		return err
	}
	if err := o.createSchema(); err != nil {
		return err
	}
	_, err := o.DB.Exec(fmt.Sprintf("set search_path='%s'", o.searchPath))
	return err
}

// CreateTables create tables for a schema.
func (o *ORM) CreateTables(schema *Schema) error {
	for _, statement := range CreateTablesStatement(schema) {
		o.logger.WithValues("statement", statement).Info("Creating table.")
		_, err := o.DB.Query(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

// connect with the database, instantiate the connection.
func (o *ORM) connect(dbname, searchPath string) error {
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s search_path=%s",
		o.config.Username,
		o.config.Password,
		dbname,
		searchPath,
	)
	if o.config.Options != "" {
		connStr = fmt.Sprintf("%s %s", connStr, o.config.Options)
	}
	var err error
	o.DB, err = sql.Open(driverName, connStr)
	return err
}

// interpolate table columne's argument with cached primary-keys, in order to complete the desired
// amount of columns with foreign-keys.
func (o *ORM) interpolate(table *Table, arguments List, cachedIDs map[string]int64) (List, error) {
	argumentWithFK := make(List, 0)
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

// resultMatrix build a matrix of results from sql.Rows.
func (o *ORM) resultMatrix(schema *Schema, rows *sql.Rows) (map[string]int, []List, error) {
	rowColumns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	// extracting row column names to create a map of name and column position
	columnIDs := map[string]int{}
	for i, name := range rowColumns {
		columnIDs[name] = i
	}

	matrix := make([]List, 0)
	// scanning row values to a single slice of slices
	for rows.Next() {
		columnValues := make(List, len(rowColumns))
		columnValuePointers := make(List, len(rowColumns))
		for i := range columnValues {
			columnValuePointers[i] = &columnValues[i]
		}
		// scanning results using pointers to populate columns slice
		if err = rows.Scan(columnValuePointers...); err != nil {
			return nil, nil, err
		}
		matrix = append(matrix, columnValues)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	return columnIDs, matrix, nil
}

// scanRows extract a matrix of results, and create a result-set object with it.
func (o *ORM) scanRows(schema *Schema, rows *sql.Rows) (*ResultSet, error) {
	// preparing row results into a single matrix of data, having back an map of column names and
	// their respective column number
	columnIDs, matrix, err := o.resultMatrix(schema, rows)
	if err != nil {
		return nil, err
	}
	return NewResultSet(schema, columnIDs, matrix)
}

// dbSelect execute a select against the schema tables using where clause and arguments informed.
// It can return errors on executing the query and building the result-set.
func (o *ORM) dbSelect(schema *Schema, where []string, arguments []interface{}) (*ResultSet, error) {
	statement := SelectStatement(schema, where)
	o.logger.WithValues("statement", statement, "where", where, "arguments", arguments).
		Info("Executing select statement")
	rows, err := o.DB.Query(statement, arguments...)
	if err != nil {
		return nil, err
	}
	return o.scanRows(schema, rows)
}

// Read namespaced-name from database, returned as a result-set instance.
// Create stores a given object in the database.
func (o *ORM) Create(schema *Schema, matrix MappedMatrix) error {
	rows := len(matrix)
	if rows == 0 {
		return fmt.Errorf("empty data informed")
	}
	logger := o.logger.WithValues("matrix-rows", rows, "schema", schema.Name)
	logger.Info("Executing create against informed schema.")

	statements := InsertStatement(schema)

	txn, err := o.DB.Begin()
	if err != nil {
		return err
	}

	tablePKCache := make(map[string]int64, len(statements))
	for i, table := range schema.Tables {
		statement := statements[i]
		arguments, found := matrix[table.Name]
		if !found {
			continue
		}
		logger = logger.WithValues(
			"statemnet", statement, "rows", len(arguments), "table", table.Name)

		// for each row found for that
		for _, argument := range arguments {
			logger.WithValues("argument", argument, "statement", statement).
				Info("Executing insert")
			// in case the case of arguments for this table being less than expected, completing the
			// slice with foreign-key cached IDs
			if len(argument) == 0 {
				continue
			}
			// completing argument with foreign-keys values, cached from previous statements
			if len(argument) < len(table.Columns)-1 {
				if argument, err = o.interpolate(table, argument, tablePKCache); err != nil {
					return err
				}
			}
			// executing insert statement and capturing primary-key
			var primaryKeyValue int64
			if err = txn.QueryRow(statement, argument...).Scan(&primaryKeyValue); err != nil {
				return err
			}
			tablePKCache[table.Name] = primaryKeyValue
		}
	}

	return txn.Commit()
}

// Read a single namespaced name from database, building back a result-set. It can return errors
// from querying the databae and building the result-set.
func (o *ORM) Read(schema *Schema, namespacedName types.NamespacedName) (*ResultSet, error) {
	metadataTable, err := schema.GetTable(fmt.Sprintf("%s_metadata", schema.Name))
	if err != nil {
		return nil, err
	}
	where := []string{
		fmt.Sprintf("%s.namespace", metadataTable.Hint),
		fmt.Sprintf("%s.name", metadataTable.Hint),
	}
	arguments := []interface{}{namespacedName.Namespace, namespacedName.Name}
	return o.dbSelect(schema, where, arguments)
}

// List all items matching labels informed. It can return errors from querying the database,
// and building a result-set with rows.
func (o *ORM) List(schema *Schema, labelsSet map[string]string) (*ResultSet, error) {
	labelsTable, err := schema.GetTable(fmt.Sprintf("%s_metadata_labels", schema.Name))
	if err != nil {
		return nil, err
	}
	where := []string{}
	arguments := []interface{}{}
	for label, value := range labelsSet {
		where = append(where, fmt.Sprintf("%s.key", labelsTable.Hint))
		where = append(where, fmt.Sprintf("%s.value", labelsTable.Hint))
		arguments = append(arguments, label)
		arguments = append(arguments, value)
	}
	return o.dbSelect(schema, where, arguments)
}

// NewORM instantiate an ORM.
func NewORM(logger logr.Logger, database string, searchPath string, config *config.Config) *ORM {
	return &ORM{
		logger: logger.WithName("orm").WithValues(
			"database", database,
			"searchPath", searchPath,
		),
		database:   database,
		searchPath: searchPath,
		config:     config,
	}
}
