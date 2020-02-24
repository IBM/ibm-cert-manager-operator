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
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CRDMap a map from the crd name to the definition of that crd
var CRDMap = map[string]*apiextensionv1beta1.CustomResourceDefinition{
	"certificates":   certificateCRD,
	"issuers":        issuerCRD,
	"clusterissuers": clusterIssuerCRD,
	"orders":         orderCRD,
	"challenges":     challengeCRD,
}

var certificateCRD = &apiextensionv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "certificates.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1beta1.NamespaceScoped,
		Names: apiextensionv1beta1.CustomResourceDefinitionNames{
			Plural:     "certificates",
			Kind:       "Certificate",
			ShortNames: []string{"cert", "certs"},
		},
		AdditionalPrinterColumns: []apiextensionv1beta1.CustomResourceColumnDefinition{
			{
				JSONPath: `.status.conditions[?(@.type=="Ready")].status`,
				Name:     "Ready",
				Type:     "string",
			},
			{
				JSONPath: ".spec.secretName",
				Name:     "Secret",
				Type:     "string",
			},
			{
				JSONPath: ".spec.issuerRef.name",
				Name:     "Issuer",
				Type:     "string",
				Priority: 1,
			},
			{
				JSONPath: `.status.conditions[?(@.type=="Ready")].message`,
				Name:     "Status",
				Type:     "string",
				Priority: 1,
			},
			{
				JSONPath:    ".metadata.creationTimestamp",
				Description: "CreationTimestamp is a timestamp representing time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists.",
				Name:        "Age",
				Type:        "date",
			},
			{
				JSONPath: ".status.notAfter",
				Name:     "Expiration",
				Type:     "string",
			},
		},
	},
}

var issuerCRD = &apiextensionv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "issuers.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1beta1.NamespaceScoped,
		Names: apiextensionv1beta1.CustomResourceDefinitionNames{
			Plural: "issuers",
			Kind:   "Issuer",
		},
	},
}

var clusterIssuerCRD = &apiextensionv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "clusterissuers.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1beta1.ClusterScoped,
		Names: apiextensionv1beta1.CustomResourceDefinitionNames{
			Plural: "clusterissuers",
			Kind:   "ClusterIssuer",
		},
	},
}

var orderCRD = &apiextensionv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "orders.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1beta1.NamespaceScoped,
		Names: apiextensionv1beta1.CustomResourceDefinitionNames{
			Plural: "orders",
			Kind:   "Order",
		},
	},
}

var challengeCRD = &apiextensionv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "challenges.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1beta1.NamespaceScoped,
		Names: apiextensionv1beta1.CustomResourceDefinitionNames{
			Plural: "challenges",
			Kind:   "Challenge",
		},
	},
}
