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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
var PodAnnotations = map[string]string{"openshift.io/scc": "restricted", "productName": "IBM Cloud Platform Common Services", "productID": "068a62892a1e4db39641342e592daa25", "productVersion": "3.4.0", "productMetric": "FREE"}

var securityAnnotationWebhook = map[string]string{"openshift.io/scc": "hostnetwork",
	"productName":    "IBM Cloud Platform Common Services",
	"productID":      "068a62892a1e4db39641342e592daa25",
	"productVersion": "3.4.0",
	"productMetric":  "FREE",
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

// DeployNamespace is the namespace the cert-manager services will be deployed in
const DeployNamespace = "ibm-common-services"
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
const ControllerImageVersion = "0.10.5"

// WebhookImageVersion is the default image version used for the cert-manager-webhook
const WebhookImageVersion = "0.10.5"

// ConfigmapWatcherVersion is the default image version used for the configmap-watcher
const ConfigmapWatcherVersion = "3.3.2"

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

// ControllerTagEnvVar is the env variable name defined in operator container for Controller Image Tag/SHA. Check operator.yaml
const ControllerTagEnvVar = "CONTROLLER_IMAGE_TAG_OR_SHA"

// WebhookTagEnvVar is the env variable name defined in operator container for Webhook Image Tag/SHA. Check operator.yaml
const WebhookTagEnvVar = "WEBHOOK_IMAGE_TAG_OR_SHA"

// CaInjectorTagEnvVar is the env variable name defined in operator container for cainjector Image Tag/SHA. Check operator.yaml
const CaInjectorTagEnvVar = "CAINJECTOR_IMAGE_TAG_OR_SHA"

// AcmeSolverTagEnvVar is the env variable name defined in operator container for acme-solver Image Tag/SHA. Check operator.yaml
const AcmeSolverTagEnvVar = "ACMESOLVER_IMAGE_TAG_OR_SHA"

// ConfigMapWatcherTagEnvVar is the env variable name defined in operator container for ConfigMap Watcher Image Tag/SHA. Check operator.yaml
const ConfigMapWatcherTagEnvVar = "CONFIGMAP_WATCHER_IMAGE_TAG_OR_SHA"

// DefaultImagePostfix is set to empty. It indicates any platform suffix that you can append to an image tag
const DefaultImagePostfix = ""

var controllerImage = GetImageID(ImageRegistry, ControllerImageName, ControllerImageVersion, DefaultImagePostfix, ControllerTagEnvVar)
var acmesolverImage = GetImageID(ImageRegistry, AcmesolverImageName, ControllerImageVersion, DefaultImagePostfix, AcmeSolverTagEnvVar)
var cainjectorImage = GetImageID(ImageRegistry, CainjectorImageName, ControllerImageVersion, DefaultImagePostfix, CaInjectorTagEnvVar)
var webhookImage = GetImageID(ImageRegistry, WebhookImageName, WebhookImageVersion, DefaultImagePostfix, WebhookTagEnvVar)
var configmapWatcherImage = GetImageID(ImageRegistry, ConfigmapWatcherImageName, ConfigmapWatcherVersion, DefaultImagePostfix, ConfigMapWatcherTagEnvVar)

// ServiceAccount is the name of the default service account to be used by cert-manager services
const ServiceAccount = "cert-manager"

// ClusterRoleName is the default name of the clusterrole and clusterrolebinding used by the cert-manager services
const ClusterRoleName = "cert-manager"

// SecurityContext values
var runAsNonRoot = true

// Liveness/Readiness Probe
var initialDelaySecondsLiveness int32 = 30
var timeoutSecondsLiveness int32 = 5
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

var initialDelaySecondsReadiness int32 = 10
var timeoutSecondsReadiness int32 = 2
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
const ResourceNS = "--cluster-resource-namespace=ibm-common-services"

const leaderElectNS = "--leader-election-namespace=cert-manager"

// AcmeSolverArg is the acme solver image to use for the cert-manager-controller
var AcmeSolverArg = "--acme-http01-solver-image=" + acmesolverImage

const webhookNSArg = "--webhook-namespace=" + DeployNamespace
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
