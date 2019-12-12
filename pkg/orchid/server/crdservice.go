package server

import (
	"github.com/isutton/orchid/pkg/orchid/runtime"
)

// crdService is a concrete implementation of CRDService.
type crdService struct {
}

// Create builds the underlying storage for object.
func (c *crdService) Create(object runtime.Object) {
	panic("implement me")
}

// NewCRDService returns a concrete implementation of CRDService.
func NewCRDService() CRDService {
	return &crdService{}
}
