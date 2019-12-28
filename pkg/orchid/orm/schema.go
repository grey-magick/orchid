package orm

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
)

// Schema around a given CR (Custom Resource), as in the group of tables required to store CR's
// payload. It also handles JSON-Schema properties to generate additional tables and columns.
type Schema struct {
	logger logr.Logger // logger instance
	Name   string      // primary table and schema name
	Tables []*Table    // schema tables
}

// Relationship describes what a new table needs to have in order to establish a relationship
type Relationship struct {
	Path        []string      // data path in respective to original object
	Constraints []*Constraint // list of constraints
	OneToMany   bool          // one-to-many flag
	KV          bool          // key-value flag
}

// HasOneToMany checks if a table matching fieldPath is one-to-many
func (s *Schema) HasOneToMany(fieldPath []string) bool {
	table := s.GetTableByPath(fieldPath)
	if table == nil {
		return false
	}
	return table.OneToMany
}

// IsKV checks if a table matching fieldPath is key-value.
func (s *Schema) IsKV(fieldPath []string) bool {
	table := s.GetTableByPath(fieldPath)
	if table == nil {
		return false
	}
	return table.KV
}

// GetHint based on table's name return its hint.
func (s *Schema) GetHint(tableName string) string {
	table, err := s.GetTable(tableName)
	if err != nil {
		return ""
	}
	return table.Hint
}

// TableName append suffix on schema name, creating an all lowercase name.
func (s *Schema) TableName(suffix string) string {
	name := strings.ReplaceAll(s.Name, ".", "_")
	return strings.ToLower(fmt.Sprintf("%s_%s", name, suffix))
}

// prepend new tables on the beggining of the slice. Reverse order helps to deal with the creation
// of tables that are dependant on each other.
func (s *Schema) prepend(table *Table) {
	s.Tables = append(s.Tables, table)
	copy(s.Tables[1:], s.Tables)
	s.Tables[0] = table
}

// TableFactory return existing or create new table instance, where when one-to-many flag true it
// appends the table instead of prepending. The sequence of tables matters during table creation
// and insertion of data.
func (s *Schema) TableFactory(tableName string, appendTable bool) *Table {
	for _, table := range s.Tables {
		if tableName == table.Name {
			return table
		}
	}

	table := NewTable(tableName)
	if appendTable {
		s.Tables = append(s.Tables, table)
	} else {
		s.prepend(table)
	}
	return table
}

// TablesReversed reverse list of tables in Schema. Useful on creating sequence of SQL statements
// on the correct dependency sequence, therefore Schema also leverages "append" vs. "prepend" of
// tables being included on schema.
func (s *Schema) TablesReversed() []*Table {
	reversed := make([]*Table, len(s.Tables))
	copy(reversed, s.Tables)
	for i := len(reversed)/2 - 1; i >= 0; i-- {
		opposite := len(reversed) - 1 - i
		reversed[i], reversed[opposite] = reversed[opposite], reversed[i]
	}
	return reversed
}

// GetTable returns a table instance, if exists.
func (s *Schema) GetTable(tableName string) (*Table, error) {
	tableName = strings.ToLower(tableName)
	for _, table := range s.Tables {
		if tableName == table.Name {
			return table, nil
		}
	}
	return nil, fmt.Errorf("unable to find table '%s' in schema", tableName)
}

// GetTableByPath returns the table instance having informed path, or nil.
func (s *Schema) GetTableByPath(fieldPath []string) *Table {
	for _, table := range s.Tables {
		if len(fieldPath) != len(table.Path) {
			continue
		}
		equals := true
		for i, field := range fieldPath {
			if table.Path[i] != field {
				equals = false
				break
			}
		}
		if !equals {
			continue
		}
		return table
	}
	return nil
}

// OneToManyTables return a slice of table names that are having one-to-many relationship.
func (s *Schema) OneToManyTables(tableName string) []string {
	tables := []string{}
	for _, table := range s.Tables {
		if tableName == table.Name {
			continue
		}

		for _, constraint := range table.ForeignKeys() {
			if constraint.ColumnName != tableName {
				continue
			}
			if constraint.RelatedTableName != tableName {
				continue
			}
			tables = append(tables, table.Name)
		}
	}
	return tables
}

// Generate trigger generation of schema for given OpenAPIV3Schema object, adding on object a new
// attribute to hold object's metadata.
func (s *Schema) Generate(openAPIV3Schema *extv1.JSONSchemaProps) error {
	openAPIV3Schema.Properties["metadata"] = jsc.MetaV1ObjectMetaOpenAPIV3Schema()

	parser := NewParser(s.logger, s)
	return parser.Parse(s.Name, Relationship{}, openAPIV3Schema)
}

// NewSchema instantiate new Schema.
func NewSchema(logger logr.Logger, name string) *Schema {
	return &Schema{
		logger: logger.WithName("schema").WithName(name),
		Name:   name,
	}
}
