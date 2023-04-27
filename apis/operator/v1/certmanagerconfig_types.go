/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:validation:XPreserveUnknownFields

// CertManagerConfigSpec defines the desired state of CertManager
type CertManagerConfigSpec struct {
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

	// +optional
	License LicenseAcceptance `json:"license,omitempty"`
}

// LicenseAcceptance defines the license specification in CSV
type LicenseAcceptance struct {
	// Accepting the license - URL: https://ibm.biz/integration-licenses
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	// +optional
	Accept bool `json:"accept"`
	// The type of license being accepted.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	Use string `json:"use,omitempty"`
	// The license being accepted where the component has multiple.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	License string `json:"license,omitempty"`
	// The license key for this deployment.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	Key string `json:"key,omitempty"`
}

//CertManagerContainerSpec defines the spec related to individual operand containers
type CertManagerContainerSpec struct {
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type CACertificate struct {
	CertName  string `json:"certName"`
	Namespace string `json:"namespace"`
}

// CertManagerConfigStatus defines the observed state of CertManagerConfig
type CertManagerConfigStatus struct {
	// It will be as "OK when all objects are created successfully
	// TODO: convert these markers for spec descriptor
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="CertManagerConfig Status"
	OverallStatus string `json:"certManagerConfigStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=certmanagerconfigs,scope=Cluster

// CertManagerConfig is the Schema for the certmanagerconfigs API. Documentation For additional details regarding install parameters check: https://ibm.biz/icpfs39install. License By installing this product you accept the license terms https://ibm.biz/icpfs39license.
type CertManagerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertManagerConfigSpec   `json:"spec,omitempty"`
	Status CertManagerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CertManagerList contains a list of CertManager
type CertManagerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertManagerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertManagerConfig{}, &CertManagerConfigList{})
}
