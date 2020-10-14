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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CertManagerSpec defines the desired state of CertManager
type CertManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ImageRegistry defines the registry where the cert-manager images should be pulled from
	ImageRegistry string `json:"imageRegistry,omitempty"`
	// ImagePostFix defines the postfix string in the image name/tag
	ImagePostFix string `json:"imagePostFix,omitempty"`
	// Webhook defines a flag which can be set/unset to deploy cert-manager webhook.
	Webhook bool `json:"enableWebhook,omitempty"`
	// ResourceNS defines the namespace where cluster-scoped cert-manager resources exist
	ResourceNS string `json:"resourceNamespace,omitempty"`
	// DisableHostNetwork defines a flag that you can set/unset if you want cert-manager to use the node network namespace. Set to true i.e. disabled by default and recommended
	DisableHostNetwork *bool `json:"disableHostNetwork,omitempty"`
	// Version defines the cert-manager-operator version
	Version string `json:"version,omitempty"`
	// CertManagerController includes spec for cert-manager-controller workload
	CertManagerController CertManagerContainerSpec `json:"certManagerController,omitempty"`
	// CertManagerWebhook includes spec for cert-manager-webhook workload
	CertManagerWebhook CertManagerContainerSpec `json:"certManagerWebhook,omitempty"`
	// CertManagerCAInjector includes spec for cert-manager-cainjector workload
	CertManagerCAInjector CertManagerContainerSpec `json:"certManagerCAInjector,omitempty"`
	// ConfigMapWatcher includes spec for icp-configmap-watcher workload
	ConfigMapWatcher CertManagerContainerSpec `json:"configMapWatcher,omitempty"`
}

//CertManagerContainerSpec defines the spec related to individual operand containers
type CertManagerContainerSpec struct {
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// CertManagerStatus defines the observed state of CertManager
type CertManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// It will be as "OK when all objects are created successfully
	OverallStatus string `json:"certManagerStatus"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CertManager is the Schema for the certmanagers API
type CertManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertManagerSpec   `json:"spec,omitempty"`
	Status CertManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertManagerList contains a list of CertManager
type CertManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertManager{}, &CertManagerList{})
}
