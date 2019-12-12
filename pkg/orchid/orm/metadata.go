package orm

// Metadata represents the extra tables needed for CRD metadata.
type Metadata struct {
	schema *Schema // schema instance
}

const (
	// omSuffix ObjectMeta
	omSuffix = "object_meta"
	// omLabelsSuffix ObjectMeta.Labels
	omLabelsSuffix = "object_meta_labels"
	// omAnnotationsSuffix ObjectMeta.Annotations
	omAnnotationsSuffix = "object_meta_annotations"
	// omOwnerReferencesSuffix ObjectMeta.OwnerReferences
	omOwnerReferencesSuffix = "object_meta_owner_references"
	// omManagedFieldsSuffix ObjectMeta.ManagedFields
	omManagedFieldsSuffix = "object_meta_managed_fields"
)

// objectMetaTable create the table refering to ObjectMeta CR entry. The ObjectMeta type is
// described at https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta.
func (m *Metadata) objectMetaTable() {
	table := m.schema.TableFactory(m.schema.TableName(omSuffix), nil)
	table.AddSerialPK()

	table.AddColumn(&Column{Name: "name", Type: PgTypeText})
	table.AddColumn(&Column{Name: "generate_name", Type: PgTypeText})
	table.AddColumn(&Column{Name: "namespace", Type: PgTypeText})
	table.AddColumn(&Column{Name: "self_link", Type: PgTypeText})
	table.AddColumn(&Column{Name: "uid", Type: PgTypeText})
	table.AddColumn(&Column{Name: "resource_version", Type: PgTypeText})
	table.AddColumn(&Column{Name: "generation", Type: PgTypeBigInt})
	table.AddColumn(&Column{Name: "creation_timestamp", Type: PgTypeText})
	table.AddColumn(&Column{Name: "deletion_timestamp", Type: PgTypeText})
	table.AddColumn(&Column{Name: "deletion_grace_period_seconds", Type: PgTypeBigInt})

	table.AddBigIntFK("labels_id", m.schema.TableName(omLabelsSuffix), false)
	table.AddBigIntFK("annotations_id", m.schema.TableName(omAnnotationsSuffix), false)
	table.AddBigIntFK("owner_references_id", m.schema.TableName(omOwnerReferencesSuffix), false)

	table.AddColumn(&Column{Name: "finalizers", Type: PgTypeTextArray})
	table.AddColumn(&Column{Name: "cluster_name", Type: PgTypeText})

	table.AddBigIntFK("managed_fields_id", m.schema.TableName(omManagedFieldsSuffix), false)
}

// objectMetaLabelsTable part of ObjectMeta, stores labels.
func (m *Metadata) objectMetaLabelsTable() {
	table := m.schema.TableFactory(m.schema.TableName(omLabelsSuffix), nil)
	table.AddBigIntPK()

	table.AddColumn(&Column{Name: "name", Type: PgTypeText})
	table.AddConstraint(&Constraint{Type: PgConstraintUnique, ColumnName: "name"})

	table.AddColumn(&Column{Name: "value", Type: PgTypeText})
}

// objectMetaAnnotationsTable part of ObjectMeta, stores annotations.
func (m *Metadata) objectMetaAnnotationsTable() {
	table := m.schema.TableFactory(m.schema.TableName(omAnnotationsSuffix), nil)
	table.AddBigIntPK()

	table.AddColumn(&Column{Name: "name", Type: PgTypeText})
	table.AddConstraint(&Constraint{Type: PgConstraintUnique, ColumnName: "name"})

	table.AddColumn(&Column{Name: "value", Type: PgTypeText})
}

// objectMetaReferencesTable part of ObjectMeta, stores references.
func (m *Metadata) objectMetaReferencesTable() {
	table := m.schema.TableFactory(m.schema.TableName(omOwnerReferencesSuffix), nil)
	table.AddBigIntPK()

	table.AddColumn(&Column{Name: "api_version", Type: PgTypeText})
	table.AddColumn(&Column{Name: "kind", Type: PgTypeText})
	table.AddColumn(&Column{Name: "name", Type: PgTypeText})
	table.AddColumn(&Column{Name: "controller", Type: PgTypeBoolean})
	table.AddColumn(&Column{Name: "block_owner_deletion", Type: PgTypeBoolean})
}

// objectMetaManagedFieldsTable part of ObjectMeta, stores managed fields.
func (m *Metadata) objectMetaManagedFieldsTable() {
	table := m.schema.TableFactory(m.schema.TableName(omManagedFieldsSuffix), nil)
	table.AddBigIntPK()

	table.AddColumn(&Column{Name: "manager", Type: PgTypeText})
	table.AddColumn(&Column{Name: "operation", Type: PgTypeText})
	table.AddColumn(&Column{Name: "api_version", Type: PgTypeText})
	table.AddColumn(&Column{Name: "time", Type: PgTypeText})
	table.AddColumn(&Column{Name: "fields_type", Type: PgTypeText})
	table.AddColumn(&Column{Name: "fields_v1", Type: PgTypeText})
}

// Add object-meta tables on informed table.
func (m *Metadata) Add(table *Table) {
	m.objectMetaTable()
	m.objectMetaAnnotationsTable()
	m.objectMetaLabelsTable()
	m.objectMetaManagedFieldsTable()
	m.objectMetaReferencesTable()

	table.AddBigIntFK("metadata_id", m.schema.TableName(omSuffix), true)
}

// NewMetadata instantiate Metadata.
func NewMetadata(schema *Schema) *Metadata {
	return &Metadata{schema: schema}
}
