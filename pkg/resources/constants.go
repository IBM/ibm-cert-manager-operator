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

var securityAnnotation = map[string]string{}

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
const DeployNamespace = "cert-manager"
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

// Default Image Values
const imageRegistry = "quay.io"

// ControllerImageVersion is the image version used for the cert-manager-controller
const ControllerImageVersion = "0.10.3"

// WebhookImageVersion is the image version used for the cert-manager-webhook
const WebhookImageVersion = "0.10.3"

// ConfigmapWatcherVersion is the image version used for the configmap-watcher
const ConfigmapWatcherVersion = "3.3.0"

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

const controllerImage = imageRegistry + "/" + ControllerImageName + ":" + ControllerImageVersion
const acmesolverImage = imageRegistry + "/" + AcmesolverImageName + ":" + ControllerImageVersion
const cainjectorImage = imageRegistry + "/" + CainjectorImageName + ":" + ControllerImageVersion
const webhookImage = imageRegistry + "/" + WebhookImageName + ":" + WebhookImageVersion
const configmapWatcherImage = imageRegistry + "/" + ConfigmapWatcherImageName + ":" + ConfigmapWatcherVersion

// ServiceAccount is the name of the default service account to be used by cert-manager services
const ServiceAccount = "default"

// ClusterRoleName is the default name of the clusterrole and clusterrolebinding used by the cert-manager services
const ClusterRoleName = "cert-manager"

// SecurityContext values
var runAsNonRoot = true
var runAsUser int64 = 10000
var fsgroup int64 = 1001

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
const webhookServingSecret = "cert-manager-webhook-tls"

const resourceNS = "--cluster-resource-namespace=kube-system"
const leaderElectNS = "--leader-election-namespace=cert-manager"
const acmeSolverArg = "--acme-http01-solver-image=" + acmesolverImage
const webhookNSArg = "--webhook-namespace=" + DeployNamespace
const webhookCASecretArg = "--webhook-ca-secret=cert-manager-webhook-ca"
const webhookServingSecretArg = "--webhook-serving-secret=" + webhookServingSecret
const webhookDNSNamesArg = "--webhook-dns-names=cert-manager-webhook,cert-manager-webhook.cert-manager,cert-manager-webhook.cert-manager.svc"

// DefaultArgs are the default arguments use for cert-manager-controller
var DefaultArgs = []string{resourceNS, leaderElectNS, webhookNSArg, webhookCASecretArg, webhookServingSecretArg, webhookDNSNamesArg}

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
