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
	corev1 "k8s.io/api/core/v1"
)

var podAffinity = &corev1.Affinity{
	NodeAffinity: &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/arch",
							Operator: "In",
							Values: []string{
								"amd64",
								"ppc64le",
								"s390x",
							},
						},
					},
				},
			},
		},
	},
}

var seccompProfile = &corev1.SeccompProfile{
	Type: corev1.SeccompProfileTypeRuntimeDefault,
}

var podSecurity = &corev1.PodSecurityContext{
	RunAsNonRoot:   &runAsNonRoot,
	SeccompProfile: seccompProfile,
}

var certManagerControllerPod = corev1.PodSpec{
	Affinity:           podAffinity,
	ServiceAccountName: "ibm-cert-manager-controller",
	SecurityContext:    podSecurity,
	Containers: []corev1.Container{
		controllerContainer,
	},
}

var certManagerWebhookPod = corev1.PodSpec{
	Affinity:           podAffinity,
	HostNetwork:        TrueVar,
	ServiceAccountName: "ibm-cert-manager-webhook",
	SecurityContext:    podSecurity,
	Containers: []corev1.Container{
		webhookContainer,
	},
}

var certManagerCainjectorPod = corev1.PodSpec{
	Affinity:           podAffinity,
	ServiceAccountName: "ibm-cert-manager-cainjector",
	SecurityContext:    podSecurity,
	Containers: []corev1.Container{
		cainjectorContainer,
	},
}
