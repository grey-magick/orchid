package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/pkg/orchid/config"
	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/test/mocks"
)

func TestAssembler_New(t *testing.T) {
	logger := klogr.New().WithName("test")
	config := &config.Config{Username: "postgres", Password: "1", Options: "sslmode=disable"}
	pgORM := orm.NewORM(logger, "postgres", "public", config)

	err := pgORM.Bootstrap()
	assert.NoError(t, err)

	schema := orm.NewSchema(logger, "assembler")

	apiSchema := mocks.OpenAPIV3SchemaMock()
	err = schema.Generate(&apiSchema)
	assert.NoError(t, err)
}
