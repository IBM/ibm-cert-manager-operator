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
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ImageRegistry      string `json:"imageRegistry,omitempty"`
	ImagePostFix       string `json:"imagePostFix,omitempty"`
	Webhook            bool   `json:"enableWebhook,omitempty"`
	ResourceNS         string `json:"resourceNamespace,omitempty"`
	DisableHostNetwork *bool  `json:"disableHostNetwork,omitempty"`
	Version            string `json:"version,omitempty"`
	//CertManagerController includes spec for cert-manager-controller workload
	CertManagerController CertManagerContainerSpec `json:"certManagerController,omitempty"`
	//CertManagerWebhook includes spec for cert-manager-webhook workload
	CertManagerWebhook CertManagerContainerSpec `json:"certManagerWebhook,omitempty"`
	//CertManagerCAInjector includes spec for cert-manager-cainjector workload
	CertManagerCAInjector CertManagerContainerSpec `json:"certManagerCAInjector,omitempty"`
	//ConfigMapWatcher includes spec for icp-configmap-watcher workload
	ConfigMapWatcher CertManagerContainerSpec `json:"configMapWatcher,omitempty"`

	//EnableCertRefresh is a flag that can be set to enable the refresh of leaf certificates based on a root CA
	EnableCertRefresh *bool `json:"enableCertRefresh,omitempty"`

	//RefreshCertsBasedOnCA is a list of CA certificate names. Leaf certificates created from the CA will be refreshed when the CA is refreshed.
	RefreshCertsBasedOnCA []CACertificate `json:"refreshCertsBasedOnCA,omitempty"`
}

type CACertificate struct {
	CertName  string `json:"certName"`
	Namespace string `json:"namespace"`
}

//CertManagerContainerSpec defines the spec related to individual operand containers
type CertManagerContainerSpec struct {
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// CertManagerStatus defines the observed state of CertManager
type CertManagerStatus struct {
	// It will be as "OK when all objects are created successfully
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="CertManager Status"
	OverallStatus string `json:"certManagerStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertManager is the Schema for the certmanagers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=certmanagers,scope=Cluster
type CertManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertManagerSpec   `json:"spec,omitempty"`
	Status CertManagerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertManagerList contains a list of CertManager
type CertManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertManager{}, &CertManagerList{})
}
