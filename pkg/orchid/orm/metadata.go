package orm

// Metadata represents the extra tables needed for CRD metadata.
type Metadata struct {
	schema *Schema // schema instance
}

const (
	// omSuffix ObjectMeta
	omSuffix = "objectMeta"
	// omLabelsSuffix ObjectMeta.Labels
	omLabelsSuffix = "objectMetaLabels"
	// omAnnotationsSuffix ObjectMeta.Annotations
	omAnnotationsSuffix = "objectMetaAnnotations"
	// omOwnerReferencesSuffix ObjectMeta.OwnerReferences
	omOwnerReferencesSuffix = "objectMetaOwnerReferences"
	// omManagedFieldsSuffix ObjectMeta.ManagedFields
	omManagedFieldsSuffix = "objectMetaManagedFields"
)

// objectMetaTable create the table refering to ObjectMeta CR entry. The ObjectMeta type is
// described at https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta.
func (m *Metadata) objectMetaTable() {
	table := m.schema.TableFactory(m.schema.TableName(omSuffix), []string{"metadata"})
	table.AddSerialPK()

	table.AddColumn(&Column{Name: "name", Type: PgTypeText, OriginalType: JSTypeString, NotNull: true})
	table.AddColumn(&Column{Name: "generateName", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "namespace", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "selfLink", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "uid", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "resourceVersion", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "generation", Type: PgTypeBigInt, OriginalType: JSTypeInteger})
	table.AddColumn(&Column{Name: "creationTimestamp", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "deletionTimestamp", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(
		&Column{Name: "deletionGracePeriodSeconds", Type: PgTypeBigInt, OriginalType: JSTypeInteger})

	// table.AddBigIntFK("labels", m.schema.TableName(omLabelsSuffix), false)
	// table.AddBigIntFK("annotations", m.schema.TableName(omAnnotationsSuffix), false)
	// table.AddBigIntFK("ownerReferences", m.schema.TableName(omOwnerReferencesSuffix), false)

	table.AddColumn(&Column{Name: "finalizers", Type: PgTypeTextArray, OriginalType: JSTypeArray})
	table.AddColumn(&Column{Name: "clusterName", Type: PgTypeText, OriginalType: JSTypeString})

	// table.AddBigIntFK("managedFields", m.schema.TableName(omManagedFieldsSuffix), false)
}

// objectMetaLabelsTable part of ObjectMeta, stores labels.
func (m *Metadata) objectMetaLabelsTable() {
	tablePath := []string{"metadata", "labels"}
	table := m.schema.TableFactory(m.schema.TableName(omLabelsSuffix), tablePath)
	table.AddBigIntPK()

	table.AddColumn(&Column{Name: "name", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddConstraint(&Constraint{Type: PgConstraintUnique, ColumnName: "name"})

	table.AddColumn(&Column{Name: "value", Type: PgTypeText, OriginalType: JSTypeString})
}

// objectMetaAnnotationsTable part of ObjectMeta, stores annotations.
func (m *Metadata) objectMetaAnnotationsTable() {
	tablePath := []string{"metadata", "annotations"}
	table := m.schema.TableFactory(m.schema.TableName(omAnnotationsSuffix), tablePath)
	table.AddBigIntPK()

	table.AddColumn(&Column{Name: "name", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddConstraint(&Constraint{Type: PgConstraintUnique, ColumnName: "name"})

	table.AddColumn(&Column{Name: "value", Type: PgTypeText, OriginalType: JSTypeString})
}

// objectMetaOwnerReferencesTable part of ObjectMeta, stores references.
func (m *Metadata) objectMetaOwnerReferencesTable() {
	tablePath := []string{"metadata", "ownerReferences"}
	table := m.schema.TableFactory(m.schema.TableName(omOwnerReferencesSuffix), tablePath)
	table.AddBigIntPK()

	table.AddColumn(&Column{Name: "apiVersion", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "kind", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "name", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "controller", Type: PgTypeBoolean, OriginalType: JSTypeBoolean})
	table.AddColumn(&Column{
		Name:         "block_owner_deletion",
		Type:         PgTypeBoolean,
		OriginalType: JSTypeBoolean,
	})
}

// objectMetaManagedFieldsTable part of ObjectMeta, stores managed fields.
func (m *Metadata) objectMetaManagedFieldsTable() {
	tablePath := []string{"metadata", "managedFields"}
	table := m.schema.TableFactory(m.schema.TableName(omManagedFieldsSuffix), tablePath)
	table.AddBigIntPK()

	// table.AddColumn(&Column{Name: "manager", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "operation", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "apiVersion", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "time", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "fieldsType", Type: PgTypeText, OriginalType: JSTypeString})
	table.AddColumn(&Column{Name: "fieldsV1", Type: PgTypeText, OriginalType: JSTypeString})
}

// Add object-meta tables on informed table.
func (m *Metadata) Add(table *Table) {
	m.objectMetaTable()
	// m.objectMetaAnnotationsTable()
	// m.objectMetaLabelsTable()
	// m.objectMetaManagedFieldsTable()
	// m.objectMetaOwnerReferencesTable()

	table.AddBigIntFK("metadata_id", m.schema.TableName(omSuffix), true)
}

// NewMetadata instantiate Metadata.
func NewMetadata(schema *Schema) *Metadata {
	return &Metadata{schema: schema}
}
