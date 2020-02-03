package resources

import (
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var CRDMap = map[string]*apiextensionv1beta1.CustomResourceDefinition{"certificates": certificateCRD, "issuers": issuerCRD, "clusterissuers": clusterIssuerCRD, "orders": orderCRD, "challenges": challengeCRD}

var certificateCRD = &apiextensionv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "certificates.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1beta1.NamespaceScoped,
		Names: apiextensionv1beta1.CustomResourceDefinitionNames{
			Plural: "certificates",
			Kind:   "Certificate",
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
		Scope:   apiextensionv1beta1.ClusterScoped,
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
		Scope:   apiextensionv1beta1.ClusterScoped,
		Names: apiextensionv1beta1.CustomResourceDefinitionNames{
			Plural: "challenges",
			Kind:   "Challenge",
		},
	},
}
