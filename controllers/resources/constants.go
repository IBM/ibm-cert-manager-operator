//
// Copyright 2022 IBM Corporation
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
var timeoutSecondsWebhook int32 = 10

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

// PodAnnotations are the annotations required for a pod
var PodAnnotations = map[string]string{"openshift.io/scc": "restricted", "productName": "IBM Cloud Platform Common Services", "productID": "068a62892a1e4db39641342e592daa25", "productMetric": "FREE"}

var securityAnnotationWebhook = map[string]string{"openshift.io/scc": "hostnetwork",
	"productName":   "IBM Cloud Platform Common Services",
	"productID":     "068a62892a1e4db39641342e592daa25",
	"productMetric": "FREE",
}

// ControllerLabels is a string of the cert-manager-controller's labels
const ControllerLabels = "app=ibm-cert-manager-controller"

// WebhookLabels is a string of the cert-manager-webhook's labels
const WebhookLabels = "app=ibm-cert-manager-webhook"

// CainjectorLabels is a string of the cert-manager-cainjector's labels
const CainjectorLabels = "app=ibm-cert-manager-cainjector"

// SecretWatchLabel is a string of secrets that watched by cert manager operator labels
const SecretWatchLabel string = "operator.ibm.com/watched-by-cert-manager"

// DefaultNamespace is the namespace the cert-manager services will be deployed in if the operator is deployed in all namespaces or locally
const DefaultNamespace = "ibm-cert-manager"

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
const ImageRegistry = "icr.io/cpopen/cpfs"

// ControllerImageVersion is the default image version used for the cert-manager-controller
const ControllerImageVersion = "0.12.0"

// WebhookImageVersion is the default image version used for the cert-manager-webhook
const WebhookImageVersion = "0.12.0"

// ControllerImageName is the image name of the cert-manager-controller
const ControllerImageName = "icp-cert-manager-controller"

// AcmesolverImageName is the image name of the cert-manager-acmesolver
const AcmesolverImageName = "icp-cert-manager-acmesolver"

// CainjectorImageName is the image name of the cert-manager-cainjector
const CainjectorImageName = "icp-cert-manager-cainjector"

// WebhookImageName is the image name of the cert-manager-webhook
const WebhookImageName = "icp-cert-manager-webhook"

// ControllerImageEnvVar is the env variable name defined in operator container for Controller Image. Check operator.yaml
const ControllerImageEnvVar = "ICP_CERT_MANAGER_CONTROLLER_IMAGE"

// WebhookImageEnvVar is the env variable name defined in operator container for Webhook Image. Check operator.yaml
const WebhookImageEnvVar = "ICP_CERT_MANAGER_WEBHOOK_IMAGE"

// CaInjectorImageEnvVar is the env variable name defined in operator container for cainjector Image. Check operator.yaml
const CaInjectorImageEnvVar = "ICP_CERT_MANAGER_CAINJECTOR_IMAGE"

// AcmeSolverImageEnvVar is the env variable name defined in operator container for acme-solver Image. Check operator.yaml
const AcmeSolverImageEnvVar = "ICP_CERT_MANAGER_ACMESOLVER_IMAGE"

// DefaultImagePostfix is set to empty. It indicates any platform suffix that you can append to an image tag
const DefaultImagePostfix = ""

var controllerImage = GetImageID(ImageRegistry, ControllerImageName, ControllerImageVersion, DefaultImagePostfix, ControllerImageEnvVar)
var acmesolverImage = GetImageID(ImageRegistry, AcmesolverImageName, ControllerImageVersion, DefaultImagePostfix, AcmeSolverImageEnvVar)
var cainjectorImage = GetImageID(ImageRegistry, CainjectorImageName, ControllerImageVersion, DefaultImagePostfix, CaInjectorImageEnvVar)
var webhookImage = GetImageID(ImageRegistry, WebhookImageName, WebhookImageVersion, DefaultImagePostfix, WebhookImageEnvVar)

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

// Cert-manager args

// WebhookServingSecret is the name of tls secret used for serving the cert-manager-webhook
const WebhookServingSecret = "cert-manager-webhook-ca"

// ResourceNS is the resource namespace arg for cert-manager-controller
var ResourceNS = "--cluster-resource-namespace=" + DeployNamespace

const leaderElectNS = "--leader-election-namespace=cert-manager"

// AcmeSolverArg is the acme solver image to use for the cert-manager-controller
var AcmeSolverArg = "--acme-http01-solver-image=" + acmesolverImage

// DefaultArgs are the default arguments use for cert-manager-controller
var DefaultArgs = []string{}
