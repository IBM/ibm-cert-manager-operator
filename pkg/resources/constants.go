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

// General
var TrueVar = true
var FalseVar = false

// CPU quantities
var cpu100 = resource.NewMilliQuantity(100, resource.DecimalSI)   // 100m
var cpu500 = resource.NewMilliQuantity(500, resource.DecimalSI)   // 500m
var cpu1000 = resource.NewMilliQuantity(1000, resource.DecimalSI) // 1000m

// Memory quantities
var memory100 = resource.NewQuantity(100*1024*1024, resource.BinarySI) // 100Mi
var memory300 = resource.NewQuantity(300*1024*1024, resource.BinarySI) // 300Mi
var memory500 = resource.NewQuantity(500*1024*1024, resource.BinarySI) // 500Mi

//
var replicaCount int32 = 1

const certManagerComponentName = "cert-manager"

// Labels
var ControllerLabelMap = map[string]string{"app": "ibm-cert-manager-controller", "app.kubernetes.io/name": "ibm-cert-manager-controller", "app.kubernetes.io/component": certManagerComponentName,
	"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": certManagerComponentName, "release": certManagerComponentName}

var WebhookLabelMap = map[string]string{"app": "ibm-cert-manager-webhook", "app.kubernetes.io/name": "ibm-cert-manager-webhook", "app.kubernetes.io/component": certManagerComponentName,
	"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": certManagerComponentName, "release": certManagerComponentName}
var CainjectorLabelMap = map[string]string{"app": "ibm-cert-manager-cainjector", "app.kubernetes.io/name": "ibm-cert-manager-cainjector", "app.kubernetes.io/component": certManagerComponentName,
	"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": certManagerComponentName, "release": certManagerComponentName}

const ControllerLabels = "app=ibm-cert-manager-controller"
const WebhookLabels = "app=ibm-cert-manager-webhook"
const CainjectorLabels = "app=ibm-cert-manager-cainjector"

const DeployNamespace = "cert-manager"
const pullPolicy = v1.PullIfNotPresent

// Container/Pod/Deployment names
const CertManagerControllerName = "cert-manager-controller"
const CertManagerAcmeSolverName = "cert-manager-acmesolver"
const CertManagerCainjectorName = "cert-manager-cainjector"
const CertManagerWebhookName = "cert-manager-webhook"

// Default Image Values
const imageRegistry = "quay.io"
const ImageVersion = "0.10.0"

const ControllerImageName = "icp-cert-manager-controller"
const AcmesolverImageName = "icp-cert-manager-acmesolver"
const CainjectorImageName = "icp-cert-manager-cainjector"
const WebhookImageName = "icp-cert-manager-webhook"

const controllerImage = imageRegistry + "/" + ControllerImageName + ":" + ImageVersion
const acmesolverImage = imageRegistry + "/" + AcmesolverImageName + ":" + ImageVersion
const cainjectorImage = imageRegistry + "/" + CainjectorImageName + ":" + ImageVersion
const webhookImage = imageRegistry + "/" + WebhookImageName + ":" + ImageVersion

const ImagePullSecret = "image-pull-secret"

// RBAC constants
const ServiceAccount = "default"
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

var DefaultArgs = []string{resourceNS, leaderElectNS, webhookNS, webhookCASecret, webhookServingSecret, webhookDNSNames}

// Affinity/Tolerations

// CRDs
var CRDs = [5]string{"certificates", "issuers", "clusterissuers", "orders", "challenges"}

const GroupVersion = "certmanager.k8s.io"
const CRDVersion = "v1alpha1"

// Namespace Declaration
var NamespaceDef = &v1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: DeployNamespace,
	},
	Spec: v1.NamespaceSpec{
		Finalizers: []v1.FinalizerName{"kubernetes"},
	},
}

// Services
// TODO - create prereq for webhook
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
