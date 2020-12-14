module github.com/redhat-marketplace/redhat-marketplace-operator/v2/reporter

go 1.15

require (
	emperror.dev/errors v0.8.0
	github.com/cespare/xxhash v1.1.0
	github.com/go-logr/logr v0.3.0
	github.com/google/uuid v1.1.2
	github.com/google/wire v0.4.0
	github.com/gotidy/ptr v1.3.0
	github.com/imdario/mergo v0.3.11
	github.com/meirf/gopart v0.0.0-20180520194036-37e9492a85a8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.15.0
	github.com/redhat-marketplace/redhat-marketplace-operator v0.0.0-20201211175424-6b3ce5b64e99 // indirect
	github.com/redhat-marketplace/redhat-marketplace-operator/v2 v2.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20201015110737-0a7fdd3b7696
	github.com/redhat-marketplace/redhat-marketplace-operator/v2 => ../
	github.com/redhat-marketplace/redhat-marketplace-operator/v2/metering => ../metering
	github.com/redhat-marketplace/redhat-marketplace-operator/v2/test => ../test
	k8s.io/api => k8s.io/api v0.19.4
	k8s.io/client-go => k8s.io/client-go v0.19.4
)
