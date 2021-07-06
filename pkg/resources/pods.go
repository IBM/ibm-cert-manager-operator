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

var podSecurity = &corev1.PodSecurityContext{
	RunAsNonRoot: &runAsNonRoot,
}

var certManagerControllerPod = corev1.PodSpec{
	Affinity:           podAffinity,
	ServiceAccountName: ServiceAccount,
	SecurityContext:    podSecurity,
	Containers: []corev1.Container{
		controllerContainer,
	},
}

var certManagerWebhookPod = corev1.PodSpec{
	Affinity:           podAffinity,
	HostNetwork:        TrueVar,
	ServiceAccountName: ServiceAccount,
	SecurityContext:    podSecurity,
	Containers: []corev1.Container{
		webhookContainer,
	},
	Volumes: []corev1.Volume{
		{
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "cert-manager-webhook-tls",
				},
			},
		},
	},
}

var certManagerCainjectorPod = corev1.PodSpec{
	Affinity:           podAffinity,
	ServiceAccountName: ServiceAccount,
	SecurityContext:    podSecurity,
	Containers: []corev1.Container{
		cainjectorContainer,
	},
}

var configmapWatcherPod = corev1.PodSpec{
	Affinity:           podAffinity,
	ServiceAccountName: ServiceAccount,
	SecurityContext:    podSecurity,
	Containers: []corev1.Container{
		configmapWatcherContainer,
	},
}
