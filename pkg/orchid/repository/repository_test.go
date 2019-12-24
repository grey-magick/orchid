package repository

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/test/mocks"
)

func TestRepository_New(t *testing.T) {
	logger := klogr.New().WithName("test")

	pgORM := orm.NewORM(logger, "user=postgres password=1 dbname=postgres sslmode=disable")
	err := pgORM.Connect()
	assert.NoError(t, err)

	repo := NewRepository(logger, pgORM)
	assert.NotNil(t, repo)

	t.Run("Bootstrap", func(t *testing.T) {
		err := repo.Bootstrap()
		require.NoError(t, err)
	})

	t.Run("Create-CRD", func(t *testing.T) {
		crd, err := mocks.UnstructuredCRDMock()
		require.NoError(t, err)

		err = repo.Create(crd)
		require.NoError(t, err)
	})

	cr, err := mocks.UnstructuredCRMock()
	require.NoError(t, err)
	t.Logf("cr='%#v'", cr)

	gvk := cr.GetObjectKind().GroupVersionKind()

	_, err = pgORM.DB.Query("truncate table mock_v1_custom cascade")
	require.NoError(t, err)

	t.Run("Create-CR", func(t *testing.T) {
		err = repo.Create(cr)
		require.NoError(t, err)
	})

	t.Run("Read-CR", func(t *testing.T) {
		namespacedName := types.NamespacedName{
			Namespace: cr.GetNamespace(),
			Name:      cr.GetName(),
		}
		u, err := repo.Read(gvk, namespacedName)
		require.NoError(t, err)

		t.Logf("u='%#v'", u)

		t.Log("## comparing original and obtained unstructured")
		diff := deep.Equal(cr, u)
		for _, entry := range diff {
			t.Log(entry)
		}
	})

	t.Run("List-CR", func(t *testing.T) {
		err = repo.Create(cr)
		require.NoError(t, err)

		options := metav1.ListOptions{LabelSelector: "label=label"}
		list, err := repo.List(gvk, options)
		require.NoError(t, err)

		t.Logf("List size '%d'", len(list.Items))
		for _, item := range list.Items {
			t.Logf("item='%#v'", item.Object)
		}

		assert.Len(t, list.Items, 2)
	})
}
