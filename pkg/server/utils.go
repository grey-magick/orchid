package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/runtime"
)

// Vars is equivalent to mux.Vars.
//
// Don't know what I'm doing here yet. This is meant to contain information encoded in the route
// such as group, version, resource and object name
type Vars map[string]string

// GetAPIVersion returns the apiVersion encoded in v.
func (v Vars) GetAPIVersion() (string, error) {
	group, ok := v["group"]
	if !ok {
		return "", errors.New("group not found")
	}

	version, ok := v["version"]
	if !ok {
		return "", errors.New("version not found")
	}

	return group + "/" + version, nil
}

// ResourceFunc maps vars to runtime.Object
type ResourceFunc func(vars Vars) runtime.Object

// Adapt decorates a ResourceFunc returning a HandlerFunc to be installed in the router.
func Adapt(resourceFunc ResourceFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// execute the given resourceFunc
		obj := resourceFunc(mux.Vars(r))
		if obj == nil {
			w.WriteHeader(404)
			return
		}

		// TODO: implement proper serialization
		jsonObj, err := json.Marshal(obj)
		if err != nil {
			// TODO: error treatment
		}

		// TODO: implement proper content type handling
		w.Header().Add("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonObj)
		if err != nil {
			// TODO: error treatment
		}

	}
}
