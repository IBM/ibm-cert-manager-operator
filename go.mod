module github.com/ibm/ibm-cert-manager-operator

go 1.16

require (
	github.com/IBM/ibm-secretshare-operator v1.11.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/pkg/errors v0.9.1
	golang.org/x/mod v0.4.2
	k8s.io/api v0.22.1
	k8s.io/apiextensions-apiserver v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/kube-aggregator v0.17.3
	sigs.k8s.io/controller-runtime v0.10.0
)
