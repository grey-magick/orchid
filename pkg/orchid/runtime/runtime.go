package runtime

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// JSONSchemaNode represents a JSON schema node.
type JSONSchemaNode map[string]interface{}

// JSONSchemaFields contains root level user defined JSON schema fields present in the DefaultObject, such
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
		currentValue := keysAndValues[i+1]
		var value JSONSchemaNode
		switch y := currentValue.(type) {
		case map[string]interface{}:
			value = y
		case JSONSchemaNode:
			value = y
		}

		f[k] = value
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
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() runtime.Object
}

// DefaultObject is the concrete Object implementation used in Orchid to represent user defined objects.
type DefaultObject struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	metav1.TypeMeta   `json:",omitempty"`
	JSONSchemaFields  JSONSchemaFields `json:",omitempty"`
}

func (o *DefaultObject) GetJSONSchemaField(name string) JSONSchemaNode {
	return o.JSONSchemaFields.GetJSONSchemaField(name)
}

func (o *DefaultObject) SetJSONSchemaField(name string, node JSONSchemaNode) {
	o.JSONSchemaFields.SetJSONSchemaField(name, node)
}

// SetJSONSchemaFields stores the given fields.
func (o *DefaultObject) SetJSONSchemaFields(fields JSONSchemaFields) {
	// TODO: DeepCopy
	o.JSONSchemaFields = fields
}

func (o *DefaultObject) DeepCopyObject() runtime.Object {
	// TODO: DeepCopy
	return o
}

var _ Object = &DefaultObject{}

func (o *DefaultObject) UnmarshalJSON(data []byte) error {
	d := make(map[string]interface{})
	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}

	// remove fields handled by ObjectMeta and TypeMeta
	delete(d, "apiVersion")
	delete(d, "kind")
	delete(d, "metadata")

	// TODO: revisit this code, it looks weird now I'm thinking
	fields := NewJSONSchemaFields(MapToList(d)...)
	o.SetJSONSchemaFields(fields)

	// create a type alias to DefaultObject to avoid deep recursion.
	type alias DefaultObject
	aux := &struct {
		*alias
	}{
		alias: (*alias)(o),
	}

	// unmarshal using the alias, which should change the DefaultObject as side effect.
	err = json.Unmarshal(data, aux)
	if err != nil {
		return err
	}

	return nil
}

// MapToList transforms a map to a slice of interface{}
func MapToList(d map[string]interface{}) []interface{} {
	var l []interface{}
	for k, v := range d {
		l = append(l, k, v)
	}
	return l
}

// NewObject returns a new Object.
func NewObject() Object {
	return &DefaultObject{}
}
