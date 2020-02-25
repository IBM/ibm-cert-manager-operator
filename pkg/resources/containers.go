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
	"k8s.io/apimachinery/pkg/api/resource"
)

var containerSecurityGeneral = &corev1.SecurityContext{
	RunAsNonRoot:             &runAsNonRoot,
	AllowPrivilegeEscalation: &FalseVar,
	ReadOnlyRootFilesystem:   &TrueVar,
	Privileged:               &FalseVar,
	Capabilities: &corev1.Capabilities{
		Drop: []corev1.Capability{
			"ALL",
		},
	},
}

var containerSecurityWebhook = &corev1.SecurityContext{
	RunAsNonRoot:             &runAsNonRoot,
	AllowPrivilegeEscalation: &FalseVar,
	ReadOnlyRootFilesystem:   &FalseVar,
	Privileged:               &FalseVar,
	Capabilities: &corev1.Capabilities{
		Drop: []corev1.Capability{
			"ALL",
		},
	},
}

var cpuMemory = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu500,
		corev1.ResourceMemory: *memory500},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory300},
}

var controllerContainer = corev1.Container{
	Name:            CertManagerControllerName,
	Image:           controllerImage,
	ImagePullPolicy: pullPolicy,
	Args:            []string{resourceNS, leaderElectNS, acmeSolverArg, webhookNSArg, webhookCASecretArg, webhookServingSecretArg, webhookDNSNamesArg},
	Env: []corev1.EnvVar{
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "POD_RESTART",
			Value: "true",
		},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &livenessExecActionController,
		},
		InitialDelaySeconds: initialDelaySecondsLiveness,
		TimeoutSeconds:      timeoutSecondsLiveness,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &readinessExecActionController,
		},
		InitialDelaySeconds: initialDelaySecondsReadiness,
		TimeoutSeconds:      timeoutSecondsReadiness,
	},
	SecurityContext: containerSecurityGeneral,
	Resources:       cpuMemory,
}

var webhookContainer = corev1.Container{
	Name:            CertManagerWebhookName,
	Image:           webhookImage,
	ImagePullPolicy: pullPolicy,
	Args:            []string{"--v=0", "--secure-port=1443", "--tls-cert-file=/certs/tls.crt", "--tls-private-key-file=/certs/tls.key"},
	Env: []corev1.EnvVar{
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &livenessExecActionWebhook,
		},
		InitialDelaySeconds: initialDelaySecondsLiveness,
		TimeoutSeconds:      timeoutSecondsLiveness,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &readinessExecActionWebhook,
		},
		InitialDelaySeconds: initialDelaySecondsReadiness,
		TimeoutSeconds:      timeoutSecondsReadiness,
	},
	SecurityContext: containerSecurityWebhook,
	Resources:       cpuMemory,
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "certs",
			MountPath: "/certs",
		},
	},
}

var cainjectorContainer = corev1.Container{
	Name:            CertManagerCainjectorName,
	Image:           cainjectorImage,
	ImagePullPolicy: pullPolicy,
	Env: []corev1.EnvVar{
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &livenessExecActionCainjector,
		},
		InitialDelaySeconds: initialDelaySecondsLiveness,
		TimeoutSeconds:      timeoutSecondsLiveness,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &readinessExecActionCainjector,
		},
		InitialDelaySeconds: initialDelaySecondsReadiness,
		TimeoutSeconds:      timeoutSecondsReadiness,
	},
	SecurityContext: containerSecurityGeneral,
	Resources:       cpuMemory,
}

var configmapWatcherContainer = corev1.Container{
	Name:            "configmap-watcher",
	Image:           configmapWatcherImage, //change
	ImagePullPolicy: pullPolicy,
	Args:            []string{"--v=0"},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &livenessExecActionConfigmapWatcher,
		},
		InitialDelaySeconds: initialDelaySecondsLiveness,
		TimeoutSeconds:      timeoutSecondsLiveness,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &readinessExecActionConfigmapWatcher,
		},
		InitialDelaySeconds: initialDelaySecondsReadiness,
		TimeoutSeconds:      timeoutSecondsReadiness,
	},
	SecurityContext: containerSecurityGeneral,
	Resources:       cpuMemory,
}
