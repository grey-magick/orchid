package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/test/mocks"
)

var (
	containersFieldPath     = []string{"spec", "template", "spec", "containers"}
	portsFieldPath          = append(containersFieldPath, "ports")
	specobjectMetaFiledPath = []string{"spec", "template", "metadata"}
)

func replicaSetSchema() *orm.Schema {
	schema := orm.NewSchema("replicaset")

	specTable := schema.TableFactory("spec", true)
	specTable.Path = []string{"spec"}

	templateTable := schema.TableFactory("spec.template", true)
	templateTable.Path = append(specTable.Path, "template")

	specTemplateSpecTable := schema.TableFactory("spec.template.spec", true)
	specTemplateSpecTable.Path = append(templateTable.Path, "spec")

	containersTable := schema.TableFactory("spec.template.spec.containers", true)
	containersTable.OneToMany = true
	containersTable.Path = append(specTemplateSpecTable.Path, "containers")

	portsTable := schema.TableFactory("spec.template.spec.containers.ports", true)
	portsTable.OneToMany = true
	portsTable.Path = append(containersTable.Path, "ports")

	return schema
}

func TestNested_New(t *testing.T) {
	schema := replicaSetSchema()

	u, err := mocks.UnstructuredReplicaSetMock()
	require.NoError(t, err)
	t.Logf("replicaSet='%+v'", u.Object)

	nested := NewNested(schema, u.Object)

	t.Run("Extract", func(t *testing.T) {
		containers, err := nested.Extract(containersFieldPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, containers)

		t.Logf("containers='%+v'", containers)
		require.Len(t, containers, 2)

		ports, err := nested.Extract(portsFieldPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, ports)

		t.Logf("ports='%+v'", ports)
		require.Len(t, ports, 4)

		names, err := nested.Extract(specobjectMetaFiledPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, names)

		t.Logf("names='%+v'", names)
		require.Len(t, names, 1)
	})
}
