//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package resources

import (
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TrueVar the variable representing the boolean value true
var TrueVar = true

// FalseVar the variable representing the boolean value false
var FalseVar = false

// CPU quantities
var cpu100 = resource.NewMilliQuantity(100, resource.DecimalSI) // 100m
var cpu500 = resource.NewMilliQuantity(500, resource.DecimalSI) // 500m

// Memory quantities
var memory300 = resource.NewQuantity(300*1024*1024, resource.BinarySI) // 300Mi
var memory500 = resource.NewQuantity(500*1024*1024, resource.BinarySI) // 500Mi

var replicaCount int32 = 1

const certManagerComponentName = "cert-manager"

// ControllerLabelMap is a map of all the labels used by cert-manager-controller
var ControllerLabelMap = map[string]string{
	"app":                          "ibm-cert-manager-controller",
	"app.kubernetes.io/name":       "ibm-cert-manager-controller",
	"app.kubernetes.io/component":  certManagerComponentName,
	"app.kubernetes.io/managed-by": "operator",
	"app.kubernetes.io/instance":   certManagerComponentName,
	"release":                      certManagerComponentName,
}

// WebhookLabelMap is a map of all the labels used by the cert-manager-webhook
var WebhookLabelMap = map[string]string{
	"app":                          "ibm-cert-manager-webhook",
	"app.kubernetes.io/name":       "ibm-cert-manager-webhook",
	"app.kubernetes.io/component":  certManagerComponentName,
	"app.kubernetes.io/managed-by": "operator",
	"app.kubernetes.io/instance":   certManagerComponentName,
	"release":                      certManagerComponentName,
	"watcher.ibm.com/opt-in":       "true",
}

// CainjectorLabelMap is a map of all the labels used by the cert-manager-cainjector
var CainjectorLabelMap = map[string]string{
	"app":                          "ibm-cert-manager-cainjector",
	"app.kubernetes.io/name":       "ibm-cert-manager-cainjector",
	"app.kubernetes.io/component":  certManagerComponentName,
	"app.kubernetes.io/managed-by": "operator",
	"app.kubernetes.io/instance":   certManagerComponentName,
	"release":                      certManagerComponentName,
}

// ConfigmapWatcherLabelMap is the labels for the configmap watcher in map format
var ConfigmapWatcherLabelMap = map[string]string{
	"app.kubernetes.io/name":       ConfigmapWatcherName,
	"app.kubernetes.io/component":  certManagerComponentName,
	"app.kubernetes.io/managed-by": "operator",
	"app.kubernetes.io/instance":   ConfigmapWatcherName,
	"release":                      certManagerComponentName,
}

// PodAnnotations are the annotations required for a pod
var PodAnnotations = map[string]string{"openshift.io/scc": "restricted", "productName": "IBM Cloud Platform Common Services", "productID": "068a62892a1e4db39641342e592daa25", "productMetric": "FREE"}

var securityAnnotationWebhook = map[string]string{"openshift.io/scc": "hostnetwork",
	"productName":   "IBM Cloud Platform Common Services",
	"productID":     "068a62892a1e4db39641342e592daa25",
	"productMetric": "FREE",
}

var webhookAnnotation = map[string]string{
	"watcher.ibm.com/configmap-resource": "kube-system/extension-apiserver-authentication",
}

// ControllerLabels is a string of the cert-manager-controller's labels
const ControllerLabels = "app=ibm-cert-manager-controller"

// WebhookLabels is a string of the cert-manager-webhook's labels
const WebhookLabels = "app=ibm-cert-manager-webhook"

// CainjectorLabels is a string of the cert-manager-cainjector's labels
const CainjectorLabels = "app=ibm-cert-manager-cainjector"

// ConfigmapWatcherLabels is a string of the configmap-watcher's labels
const ConfigmapWatcherLabels = "app.kubernetes.io/name=configmap-watcher"

// DefaultNamespace is the namespace the cert-manager services will be deployed in if the operator is deployed in all namespaces or locally
const DefaultNamespace = "ibm-common-services"

// PodNamespace is the namespace the the operator is getting deployed (set in an env var)
var PodNamespace = os.Getenv("POD_NAMESPACE")

// DeployNamespace is the namespace the cert-manager services will be deployed in
var DeployNamespace = GetDeployNamespace()

const pullPolicy = v1.PullIfNotPresent

// CertManagerControllerName is the name of the container/pod/deployment for cert-manager-controller
const CertManagerControllerName = "cert-manager-controller"

// CertManagerAcmeSolverName is the name of the container/pod/deployment for cert-manager-acmesolver
const CertManagerAcmeSolverName = "cert-manager-acmesolver"

// CertManagerCainjectorName is the name of the container/pod/deployment for cert-manager-cainjector
const CertManagerCainjectorName = "cert-manager-cainjector"

// CertManagerWebhookName is the name of the container/pod/deployment for cert-manager-webhook
const CertManagerWebhookName = "cert-manager-webhook"

// ConfigmapWatcherName is the name of the container/pod/deployment for the configmap-watcher
const ConfigmapWatcherName = "configmap-watcher"

// ImageRegistry is the default image registry for the operand deployments
const ImageRegistry = "quay.io/opencloudio"

// ControllerImageVersion is the default image version used for the cert-manager-controller
const ControllerImageVersion = "0.12.0"

// WebhookImageVersion is the default image version used for the cert-manager-webhook
const WebhookImageVersion = "0.12.0"

// ConfigmapWatcherVersion is the default image version used for the configmap-watcher
const ConfigmapWatcherVersion = "3.4.0"

// ControllerImageName is the image name of the cert-manager-controller
const ControllerImageName = "icp-cert-manager-controller"

// AcmesolverImageName is the image name of the cert-manager-acmesolver
const AcmesolverImageName = "icp-cert-manager-acmesolver"

// CainjectorImageName is the image name of the cert-manager-cainjector
const CainjectorImageName = "icp-cert-manager-cainjector"

// WebhookImageName is the image name of the cert-manager-webhook
const WebhookImageName = "icp-cert-manager-webhook"

// ConfigmapWatcherImageName is the name of the configmap watcher image
const ConfigmapWatcherImageName = "icp-configmap-watcher"

// ControllerImageEnvVar is the env variable name defined in operator container for Controller Image. Check operator.yaml
const ControllerImageEnvVar = "ICP_CERT_MANAGER_CONTROLLER_IMAGE"

// WebhookImageEnvVar is the env variable name defined in operator container for Webhook Image. Check operator.yaml
const WebhookImageEnvVar = "ICP_CERT_MANAGER_WEBHOOK_IMAGE"

// CaInjectorImageEnvVar is the env variable name defined in operator container for cainjector Image. Check operator.yaml
const CaInjectorImageEnvVar = "ICP_CERT_MANAGER_CAINJECTOR_IMAGE"

// AcmeSolverImageEnvVar is the env variable name defined in operator container for acme-solver Image. Check operator.yaml
const AcmeSolverImageEnvVar = "ICP_CERT_MANAGER_ACMESOLVER_IMAGE"

// ConfigMapWatcherImageEnvVar is the env variable name defined in operator container for ConfigMap Watcher Image. Check operator.yaml
const ConfigMapWatcherImageEnvVar = "ICP_CONFIGMAP_WATCHER_IMAGE"

// DefaultImagePostfix is set to empty. It indicates any platform suffix that you can append to an image tag
const DefaultImagePostfix = ""

var controllerImage = GetImageID(ImageRegistry, ControllerImageName, ControllerImageVersion, DefaultImagePostfix, ControllerImageEnvVar)
var acmesolverImage = GetImageID(ImageRegistry, AcmesolverImageName, ControllerImageVersion, DefaultImagePostfix, AcmeSolverImageEnvVar)
var cainjectorImage = GetImageID(ImageRegistry, CainjectorImageName, ControllerImageVersion, DefaultImagePostfix, CaInjectorImageEnvVar)
var webhookImage = GetImageID(ImageRegistry, WebhookImageName, WebhookImageVersion, DefaultImagePostfix, WebhookImageEnvVar)
var configmapWatcherImage = GetImageID(ImageRegistry, ConfigmapWatcherImageName, ConfigmapWatcherVersion, DefaultImagePostfix, ConfigMapWatcherImageEnvVar)

// ServiceAccount is the name of the default service account to be used by cert-manager services
const ServiceAccount = "cert-manager"

// ClusterRoleName is the default name of the clusterrole and clusterrolebinding used by the cert-manager services
const ClusterRoleName = "cert-manager"

// SecurityContext values
var runAsNonRoot = true

// Liveness/Readiness Probe
var initialDelaySecondsLiveness int32 = 60
var timeoutSecondsLiveness int32 = 10
var periodSecondsLiveness int32 = 30
var failureThresholdLiveness int32 = 10
var livenessExecActionController = v1.ExecAction{
	Command: []string{"sh", "-c", "pgrep cert-manager -l"},
}
var livenessExecActionCainjector = v1.ExecAction{
	Command: []string{"sh", "-c", "pgrep cainjector -l"},
}
var livenessExecActionWebhook = v1.ExecAction{
	Command: []string{"sh", "-c", "pgrep webhook -l"},
}
var livenessExecActionConfigmapWatcher = v1.ExecAction{
	Command: []string{"sh", "-c", "pgrep watcher -l"},
}

var initialDelaySecondsReadiness int32 = 60
var timeoutSecondsReadiness int32 = 10
var periodSecondsReadiness int32 = 30
var failureThresholdReadiness int32 = 10
var readinessExecActionController = v1.ExecAction{
	Command: []string{"sh", "-c", "exec echo start cert-manager"},
}
var readinessExecActionCainjector = v1.ExecAction{
	Command: []string{"sh", "-c", "exec echo start cert-manager cainjector"},
}
var readinessExecActionWebhook = v1.ExecAction{
	Command: []string{"sh", "-c", "exec echo start cert-manager webhook"},
}
var readinessExecActionConfigmapWatcher = v1.ExecAction{
	Command: []string{"sh", "-c", "exec echo start configmap-watcher"},
}

// Cert-manager args

// WebhookServingSecret is the name of tls secret used for serving the cert-manager-webhook
const WebhookServingSecret = "cert-manager-webhook-tls"

// ResourceNS is the resource namespace arg for cert-manager-controller
var ResourceNS = "--cluster-resource-namespace=" + DeployNamespace

const leaderElectNS = "--leader-election-namespace=cert-manager"

// AcmeSolverArg is the acme solver image to use for the cert-manager-controller
var AcmeSolverArg = "--acme-http01-solver-image=" + acmesolverImage

var webhookNSArg = "--webhook-namespace=" + DeployNamespace

const webhookCASecretArg = "--webhook-ca-secret=cert-manager-webhook-ca"
const webhookServingSecretArg = "--webhook-serving-secret=" + WebhookServingSecret

const webhookDNSNamesArg = "--webhook-dns-names=cert-manager-webhook,cert-manager-webhook.cert-manager,cert-manager-webhook.cert-manager.svc"
const controllersArg = "--controllers=certificates,issuers,clusterissuers,orders,challenges,webhook-bootstrap"

// DefaultArgs are the default arguments use for cert-manager-controller
var DefaultArgs = []string{webhookCASecretArg, webhookServingSecretArg, controllersArg}

// CRDs is the list of crds created/used by cert-manager in this version
var CRDs = [5]string{"certificates", "issuers", "clusterissuers", "orders", "challenges"}

// GroupVersion is the cert-manager's crd group version
const GroupVersion = "certmanager.k8s.io"

//CRDVersion is the cert-manager's crd version
const CRDVersion = "v1alpha1"

// NamespaceDef is the namespace spec for the cert-manager services and will be where the service is deployed
var NamespaceDef = &v1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: DeployNamespace,
		Labels: map[string]string{
			"certmanager.k8s.io/disable-validation": "true",
		},
	},
	Spec: v1.NamespaceSpec{
		Finalizers: []v1.FinalizerName{"kubernetes"},
	},
}

// CSCAIssuerLabelMap is the labels for the CS CA Issuer in map format
var CSCAIssuerLabelMap = map[string]string{
	"app.kubernetes.io/name":       "cert-manager",
	"app.kubernetes.io/managed-by": "ibm-cert-manager-operator",
	"app.kubernetes.io/instance":   "ibm-cert-manager-operator",
}

//CSCAIssuerName is the name of the CS CA Issuer
const CSCAIssuerName = "cs-ca-issuer"

//CSCACertName is the name of the CS CA certificate
const CSCACertName = "cs-ca-certificate"

//CSCASecretName is the name of the CA certificate secret
const CSCASecretName = "cs-ca-certificate-secret"

//RhacmNamespace is the namespace where RHACM is installed
const RhacmNamespace = "open-cluster-management"

//RhacmCRName is the RHACM CR name
const RhacmCRName = "multiclusterhub"

//RhacmSecretShareCRName is the Secret Share CR Name that copies the cs-ca-certificate-secret
var RhacmSecretShareCRName = "rhacm-cs-ca-certificate-secret-share"

//RhacmGVK identifies the RHACM CRD
var RhacmGVK = schema.GroupVersionKind{
	Group:   "operator.open-cluster-management.io",
	Kind:    "MultiClusterHub",
	Version: "v1",
}

// DefaultEnableCertRefresh is set to true
const DefaultEnableCertRefresh = true

// DefaultCANames is the default CA names for which the leaf certs need to be refreshed
var DefaultCANames = []string{"cs-ca-certificate", "mongodb-root-ca-cert"}

// CertManager instance name
const CertManagerInstanceName = "default"

// OdlmDeploymentName is the deployment name of ODLM
const OdlmDeploymentName = "operand-deployment-lifecycle-manager"

// ProductName is the name of Common Services
const ProductName = "IBM Cloud Pak Foundational Services"
