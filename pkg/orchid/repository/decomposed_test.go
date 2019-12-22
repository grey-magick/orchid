package repository

import "testing"

func Test_decoposedPath(t *testing.T) {
	schema := replicaSetSchema()

	decomposedNested := decomposePaths(schema, portsFieldPath)
	t.Logf("decomposedNested='%+v'", decomposedNested)

	decomposed := decomposePaths(schema, specobjectMetaFiledPath)
	t.Logf("decomposed='%+v'", decomposed)

}
