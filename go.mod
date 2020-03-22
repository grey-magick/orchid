module github.com/isutton/orchid

go 1.14

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-test/deep v1.0.4
	github.com/gorilla/mux v1.7.3
	github.com/lib/pq v1.2.0
	github.com/otaviof/go-sqlfmt v0.0.2
	github.com/pmezard/go-difflib v1.0.0
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.0.0-20191204090421-cd61debedab5
	k8s.io/apimachinery v0.17.4
	k8s.io/klog v1.0.0
)

replace github.com/kanmu/go-sqlfmt => github.com/otaviof/go-sqlfmt v0.0.2-0.20200322101854-ac6cb97d585a
