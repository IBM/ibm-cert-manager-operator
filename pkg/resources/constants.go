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
	"k8s.io/apimachinery/pkg/util/intstr"
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

//
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

// ControllerLabels is a string of the cert-manager-controller's labels
const ControllerLabels = "app=ibm-cert-manager-controller"

// WebhookLabels is a string of the cert-manager-webhook's labels
const WebhookLabels = "app=ibm-cert-manager-webhook"

// CainjectorLabels is a string of the cert-manager-cainjector's labels
const CainjectorLabels = "app=ibm-cert-manager-cainjector"

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

// Default Image Values
const imageRegistry = "quay.io"

// ImageVersion is the image version used for the cert-manager services
const ImageVersion = "0.10.0"

// ControllerImageName is the image name of the cert-manager-controller
const ControllerImageName = "icp-cert-manager-controller"

// AcmesolverImageName is the image name of the cert-manager-acmesolver
const AcmesolverImageName = "icp-cert-manager-acmesolver"

// CainjectorImageName is the image name of the cert-manager-cainjector
const CainjectorImageName = "icp-cert-manager-cainjector"

// WebhookImageName is the image name of the cert-manager-webhook
const WebhookImageName = "icp-cert-manager-webhook"

const controllerImage = imageRegistry + "/" + ControllerImageName + ":" + ImageVersion
const acmesolverImage = imageRegistry + "/" + AcmesolverImageName + ":" + ImageVersion
const cainjectorImage = imageRegistry + "/" + CainjectorImageName + ":" + ImageVersion
const webhookImage = imageRegistry + "/" + WebhookImageName + ":" + ImageVersion

// ImagePullSecret is the default image pull secret name
const ImagePullSecret = "image-pull-secret"

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

// Cert-manager args
const resourceNS = "--cluster-resource-namespace=kube-system"
const leaderElectNS = "--leader-election-namespace=cert-manager"
const acmeSolver = "--acme-http01-solver-image=" + acmesolverImage
const webhookNS = "--webhook-namespace=cert-manager"
const webhookCASecret = "--webhook-ca-secret=cert-manager-webhook-ca"
const webhookServingSecret = "--webhook-serving-secret=cert-manager-webhook-tls"
const webhookDNSNames = "--webhook-dns-names=cert-manager-webhook,cert-manager-webhook.cert-manager,cert-manager-webhook.cert-manager.svc"

// DefaultArgs are the default arguments use for cert-manager-controller
var DefaultArgs = []string{resourceNS, leaderElectNS, webhookNS, webhookCASecret, webhookServingSecret, webhookDNSNames}

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
	},
	Spec: v1.NamespaceSpec{
		Finalizers: []v1.FinalizerName{"kubernetes"},
	},
}

// Services
// TODO - create prereqs for webhook
var webhookSvc = &v1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "ibm-cert-manager-webhook-svc",
		Namespace: DeployNamespace,
		Labels:    WebhookLabelMap,
	},
	Spec: v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Name: "https",
				Port: 443,
				TargetPort: intstr.IntOrString{
					IntVal: 1443,
				},
			},
		},
		Selector: WebhookLabelMap,
		Type:     v1.ServiceTypeClusterIP,
	},
}
