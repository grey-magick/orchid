package repository

import (
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

	model := NewRepository(pgORM)
	assert.NotNil(t, model)

	t.Run("Bootstrap", func(t *testing.T) {
		err := model.Bootstrap()
		require.NoError(t, err)
	})

	t.Run("Create-CRD", func(t *testing.T) {
		crd, err := mocks.UnstructuredCRDMock()
		require.NoError(t, err)

		err = model.Create(crd)
		require.NoError(t, err)
	})

	t.Run("Create-CR", func(t *testing.T) {
		cr, err := mocks.UnstructuredCRMock()
		require.NoError(t, err)

		t.Logf("cr='%+v'", cr)

		err = model.Create(cr)
		assert.NoError(t, err)
	})
}
