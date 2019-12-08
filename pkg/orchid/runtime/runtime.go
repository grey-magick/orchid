package runtime

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JSONSchemaNode represents a JSON schema node.
type JSONSchemaNode map[string]interface{}

// JSONSchemaFields contains root level user defined JSON schema fields present in the object, such
// as "spec" and "status".
type JSONSchemaFields map[string]JSONSchemaNode

// NewJSONSchemaFields initializes a new value with keysAndValues.
func NewJSONSchemaFields(keysAndValues ...interface{}) JSONSchemaFields {
	keysAndValuesLen := len(keysAndValues)

	if keysAndValuesLen%2 != 0 {
		panic("should be even")
	}

	f := make(JSONSchemaFields, 2)

	for i := 0; i < keysAndValuesLen; i = i + 2 {
		k, ok := keysAndValues[i].(string)
		if !ok {
			panic("should be string")
		}
		v := keysAndValues[i+1].(JSONSchemaNode)
		f[k] = v
	}

	return f
}

// GetJSONSchemaField returns the node for the given name if any.
func (f JSONSchemaFields) GetJSONSchemaField(name string) JSONSchemaNode {
	if v, ok := f[name]; ok {
		// TODO: DeepCopy
		return v
	}
	return nil
}

// SetJSONSchemaField stores a node for the given name.
func (f JSONSchemaFields) SetJSONSchemaField(name string, node JSONSchemaNode) {
	// TODO: DeepCopy
	f[name] = node
}

// Object extends Kubernetes metav1.Object type with JSON schema capabilities.
type Object interface {
	metav1.Object
	SetJSONSchemaFields(fields JSONSchemaFields)
	GetJSONSchemaField(name string) JSONSchemaNode
	SetJSONSchemaField(name string, node JSONSchemaNode)
}

// object is the concrete Object implementation used in Orchid to represent user defined objects.
type object struct {
	metav1.ObjectMeta
	metav1.TypeMeta
	JSONSchemaFields
}

// SetJSONSchemaFields stores the given fields.
func (o *object) SetJSONSchemaFields(fields JSONSchemaFields) {
	// TODO: DeepCopy
	o.JSONSchemaFields = fields
}

var _ Object = &object{}

// NewObject returns a new Object.
func NewObject() Object {
	return &object{}
}
