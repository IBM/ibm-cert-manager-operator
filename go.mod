module github.com/IBM/ibm-cert-manager-operator

go 1.13

require (
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/kube-aggregator v0.19.2
	sigs.k8s.io/controller-runtime v0.6.2

	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0 // indirect
)
