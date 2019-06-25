module github.com/cynepco3hahue/machine-health-check-operator

require (
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/ghodss/yaml v0.0.0-20150909031657-73d445a93680
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/gomega v1.4.2 // indirect
	github.com/openshift/api v3.9.1-0.20190621203108-e6261f37404f+incompatible
	github.com/openshift/client-go v3.9.0+incompatible
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0-20190620073856-dcce3486da33
	k8s.io/apimachinery v0.0.0-20190620073744-d16981aedf33
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208 // indirect
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace github.com/spf13/cobra => github.com/spf13/cobra v0.0.3

replace github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190403184916-6209e06506a9

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70

replace k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5

replace k8s.io/apiextensions-apiserver => github.com/openshift/kubernetes-apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed

replace k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20190313205120-d7deff9243b1

replace k8s.io/client-go => github.com/openshift/kubernetes-client-go v2.0.0-alpha.0.0.20190313235726-6ee68ca5fd83+incompatible

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0-beta.1.0.20190520212815-96b67f231945

replace sigs.k8s.io/cluster-api => github.com/openshift/cluster-api v0.0.0-20190619113136-046d74a3bd91

replace github.com/openshift/cluster-api => github.com/openshift/cluster-api v0.0.0-20190619113136-046d74a3bd91
