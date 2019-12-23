package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/test/mocks"
)

func TestAssembler_New(t *testing.T) {
	logger := klogr.New().WithName("test")

	pgORM := orm.NewORM(logger, "user=postgres password=1 dbname=postgres sslmode=disable")
	err := pgORM.Connect()
	assert.NoError(t, err)

	schema := orm.NewSchema(logger, "assembler")

	apiSchema := mocks.OpenAPIV3SchemaMock()
	err = schema.GenerateCR(&apiSchema)
	assert.NoError(t, err)
}
