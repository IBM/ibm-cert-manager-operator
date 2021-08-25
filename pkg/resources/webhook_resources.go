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
	admRegv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiRegv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

var valPath = "/validate"
var mutationPath = "/mutate"
var failPolicy = admRegv1beta1.Fail
var sideEffect = admRegv1beta1.SideEffectClassNone

// MutatingWebhook is the mutating webhook definition for cert-manager-webhook
var MutatingWebhook = &admRegv1beta1.MutatingWebhookConfiguration{
	ObjectMeta: metav1.ObjectMeta{
		Name:   CertManagerWebhookName,
		Labels: WebhookLabelMap,
		Annotations: map[string]string{
			"cert-manager.io/inject-ca-from-secret": DeployNamespace + "/" + WebhookServingSecret,
		},
	},
	Webhooks: []admRegv1beta1.MutatingWebhook{
		{
			Name: "webhook.cert-manager.io",
			ClientConfig: admRegv1beta1.WebhookClientConfig{
				Service: &admRegv1beta1.ServiceReference{
					Namespace: DeployNamespace,
					Name:      CertManagerWebhookName,
					Path:      &mutationPath,
				},
			},
			Rules: []admRegv1beta1.RuleWithOperations{
				{
					Operations: []admRegv1beta1.OperationType{
						admRegv1beta1.Create,
						admRegv1beta1.Update,
					},
					Rule: admRegv1beta1.Rule{
						APIGroups: []string{
							"cert-manager.io",
							"acme.cert-manager.io",
						},
						APIVersions: []string{
							"v1",
						},
						Resources: []string{
							"*/*",
						},
					},
				},
			},
			FailurePolicy:           &failPolicy,
			SideEffects:             &sideEffect,
			AdmissionReviewVersions: []string{"v1"},
			TimeoutSeconds:          &timeoutSecondsWebhook,
		},
	},
}

//const injectSecretCA = DeployNamespace + "/" + webhookServingSecret

// APISvcName is the name used for cert-manager-webhooks' apiservice definition
const APISvcName = "v1beta1.webhook.certmanager.k8s.io"

// APIService is the apiservice for cert-manager-webhook
var APIService = &apiRegv1.APIService{
	ObjectMeta: metav1.ObjectMeta{
		Name: APISvcName,
		Labels: map[string]string{
			"app": "ibm-cert-manager-webhook",
		},
		Annotations: map[string]string{
			//"certmanager.k8s.io/inject-ca-from-secret": injectSecretCA,
		},
	},
	Spec: apiRegv1.APIServiceSpec{
		Group:                "webhook.certmanager.k8s.io",
		GroupPriorityMinimum: 1000,
		VersionPriority:      15,
		Service: &apiRegv1.ServiceReference{
			Name: CertManagerWebhookName,
			//Namespace: DeployNamespace,
		},
		Version: "v1beta1",
	},
}

// WebhookSvc is the service definition for cert-manager-webhook
var WebhookSvc = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      CertManagerWebhookName,
		Namespace: DeployNamespace,
		Labels: map[string]string{
			"app": "ibm-cert-manager-webhook",
		},
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Name: "https",
				Port: 443,
				TargetPort: intstr.IntOrString{
					IntVal: 10250,
				},
			},
		},
		Selector: map[string]string{
			"app": "ibm-cert-manager-webhook",
		},
		Type: corev1.ServiceTypeClusterIP,
	},
}

// ValidatingWebhook is the validating webhook definition for cert-manager-webhook
var ValidatingWebhook = &admRegv1beta1.ValidatingWebhookConfiguration{
	ObjectMeta: metav1.ObjectMeta{
		Name:   CertManagerWebhookName,
		Labels: WebhookLabelMap,
		Annotations: map[string]string{
			"cert-manager.io/inject-ca-from-secret": DeployNamespace + "/" + WebhookServingSecret,
		},
	},
	Webhooks: []admRegv1beta1.ValidatingWebhook{
		{
			Name: "webhook.cert-manager.io",
			Rules: []admRegv1beta1.RuleWithOperations{
				{
					Operations: []admRegv1beta1.OperationType{
						admRegv1beta1.Create,
						admRegv1beta1.Update,
					},
					Rule: admRegv1beta1.Rule{
						APIGroups: []string{
							"cert-manager.io",
							"acme.cert-manager.io",
						},
						APIVersions: []string{
							"v1",
						},
						Resources: []string{
							"*/*",
						},
					},
				},
			},
			AdmissionReviewVersions: []string{"v1"},
			ClientConfig: admRegv1beta1.WebhookClientConfig{
				Service: &admRegv1beta1.ServiceReference{
					Namespace: DeployNamespace,
					Name:      CertManagerWebhookName,
					Path:      &valPath,
				},
			},
			FailurePolicy: &failPolicy,
			SideEffects:   &sideEffect,
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "cert-manager.io/disable-validation",
						Operator: metav1.LabelSelectorOpNotIn,
						Values:   []string{"true"},
					},
					{
						Key:      "name",
						Operator: metav1.LabelSelectorOpNotIn,
						Values:   []string{DeployNamespace},
					},
				},
			},
			TimeoutSeconds: &timeoutSecondsWebhook,
		},
	},
}
