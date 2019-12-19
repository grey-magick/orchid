package repository

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/test/mocks"
)

func TestRepository_New(t *testing.T) {
	pgORM := orm.NewORM("user=postgres password=1 dbname=postgres sslmode=disable")
	err := pgORM.Connect()
	assert.NoError(t, err)

	repo := NewRepository(pgORM)
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

	t.Run("Create-CR", func(t *testing.T) {
		cr, err := mocks.UnstructuredCRMock()
		require.NoError(t, err)

		t.Logf("cr='%+v'", cr)

		err = repo.Create(cr)
		assert.NoError(t, err)

		schemaName := repo.schemaName(cr.GetObjectKind().GroupVersionKind())
		sqlLib := orm.NewSQL(repo.schemaFactory(schemaName))
		fmt.Println(sqlLib.Select())
	})
}
