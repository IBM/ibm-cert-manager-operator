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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ServiceAccountsToCreate = &corev1.ServiceAccountList{
	Items: []corev1.ServiceAccount{*ControllerServiceAccount, *CAInjectorServiceAccount},
}

var RolesToCreate = &rbacv1.RoleList{
	Items: []rbacv1.Role{*ControllerRole, *CAInjectorRole},
}

var RoleBindingsToCreate = &rbacv1.RoleBindingList{
	Items: []rbacv1.RoleBinding{*ControllerRoleBinding, *CAInjectorRoleBinding},
}

var ClusterRolesToCreate = &rbacv1.ClusterRoleList{
	Items: []rbacv1.ClusterRole{*ControllerViewClusterRole, *ControllerEditClusterRole, *ControllerApproveClusterRole, *ControllerCertificateSigningRequestsClusterRole, *ControllerIssuersClusterRole, *ControllerClusterIssuersClusterRole, *ControllerCertificatesClusterRole, *ControllerOrdersClusterRole, *ControllerChallengesClusterRole, *ControllerIngressShimClusterRole, *CAInjectorClusterRole},
}

var ClusterRoleBindingsToCreate = &rbacv1.ClusterRoleBindingList{
	Items: []rbacv1.ClusterRoleBinding{*ControllerApproveClusterRoleBinding, *ControllerCertificateSigningRequestsClusterRoleBinding, *ControllerIssuersClusterRoleBinding, *ControllerClusterIssuersClusterRoleBinding, *ControllerCertificatesClusterRoleBinding, *ControllerOrdersClusterRoleBinding, *ControllerChallengesClusterRoleBinding, *ControllerIngressShimClusterRoleBinding, *CAInjectorClusterRoleBinding},
}

var ControllerServiceAccount = &corev1.ServiceAccount{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller",
	},
}

var ControllerRole = &rbacv1.Role{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "ibm-cert-manager-controller:leaderelection",
		Namespace: "ibm-common-services",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:         []string{"get", "update", "patch"},
			APIGroups:     []string{""},
			Resources:     []string{"configmaps"},
			ResourceNames: []string{"cert-manager-controller"},
		},
		{
			Verbs:     []string{"create"},
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
		},
		{
			Verbs:         []string{"get", "update", "patch"},
			APIGroups:     []string{"coordination.k8s.io"},
			Resources:     []string{"leases"},
			ResourceNames: []string{"cert-manager-controller"},
		},
		{
			Verbs:     []string{"create"},
			APIGroups: []string{"coordination.k8s.io"},
			Resources: []string{"leases"},
		},
	},
}

var ControllerRoleBinding = &rbacv1.RoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "ibm-cert-manager-controller:leaderelection",
		Namespace: "ibm-common-services",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "ibm-cert-manager-controller:leaderelection",
	},
}

var ControllerViewClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-view",
		Labels: map[string]string{
			"rbac.authorization.k8s.io/aggregate-to-view":  "true",
			"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
			"rbac.authorization.k8s.io/aggregate-to-admin": "true",
		},
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"certificates", "certificaterequests", "issuers"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"challenges", "orders"},
		},
	},
}

var ControllerEditClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-edit",
		Labels: map[string]string{
			"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
			"rbac.authorization.k8s.io/aggregate-to-admin": "true",
		},
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"certificates", "certificaterequests", "issuers"},
		},
		{
			Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"challenges", "orders"},
		},
	},
}

var ControllerApproveClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-approve:cert-manager-io",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:         []string{"approve"},
			APIGroups:     []string{"cert-manager.io"},
			Resources:     []string{"signers"},
			ResourceNames: []string{"issuers.cert-manager.io/*", "clusterissuers.cert-manager.io/*"},
		},
	},
}

var ControllerApproveClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-approve:cert-manager-io",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-controller-approve:cert-manager-io",
	},
}

var ControllerCertificateSigningRequestsClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-certificatesigningrequests",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch", "update"},
			APIGroups: []string{"certificates.k8s.io"},
			Resources: []string{"certificatesigningrequests"},
		},
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"certificates.k8s.io"},
			Resources: []string{"certificatesigningrequests/status"},
		},
		{
			Verbs:         []string{"sign"},
			APIGroups:     []string{"certificates.k8s.io"},
			Resources:     []string{"signers"},
			ResourceNames: []string{"issuers.cert-manager.io/*", "clusterissuers.cert-manager.io/*"},
		},
		{
			Verbs:     []string{"create"},
			APIGroups: []string{"authorization.k8s.io"},
			Resources: []string{"subjectaccessreviews"},
		},
	},
}

var ControllerCertificateSigningRequestsClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-certificatesigningrequests",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-controller-certificatesigningrequests",
	},
}

var ControllerIssuersClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-issuers",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"issuers", "issuers/status"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"issuers"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
		{
			Verbs:     []string{"create", "patch"},
			APIGroups: []string{""},
			Resources: []string{"events"},
		},
	},
}

var ControllerIssuersClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-issuers",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-controller-issuers",
	},
}

var ControllerClusterIssuersClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-clusterissuers",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"clusterissuers", "clusterissuers/status"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"clusterissuers"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
		{
			Verbs:     []string{"create", "patch"},
			APIGroups: []string{""},
			Resources: []string{"events"},
		},
	},
}

var ControllerClusterIssuersClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-clusterissuers",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-controller-clusterissuers",
	},
}

var ControllerCertificatesClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-certificates",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"certificates", "certificates/status", "certificaterequests", "certificaterequests/status"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"certificates", "certificaterequests", "clusterissuers", "issuers"},
		},
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"certificates/finalizers", "certificaterequests/finalizers"},
		},
		{
			Verbs:     []string{"create", "delete", "get", "list", "watch"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"orders"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
		{
			Verbs:     []string{"create", "patch"},
			APIGroups: []string{""},
			Resources: []string{"events"},
		},
	},
}

var ControllerCertificatesClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-certificates",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-controller-certificates",
	},
}

var ControllerOrdersClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-orders",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"orders", "orders/status"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"orders", "challenges"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"clusterissuers", "issuers"},
		},
		{
			Verbs:     []string{"create", "delete"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"challenges"},
		},
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"orders/finalizers"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
		{
			Verbs:     []string{"create", "patch"},
			APIGroups: []string{""},
			Resources: []string{"events"},
		},
	},
}

var ControllerOrdersClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-orders",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-controller-orders",
	},
}

var ControllerChallengesClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-challenges",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"challenges", "challenges/status"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"challenges"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"clusterissuers", "issuers"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
		{
			Verbs:     []string{"create", "patch"},
			APIGroups: []string{""},
			Resources: []string{"events"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "create", "delete"},
			APIGroups: []string{""},
			Resources: []string{"pods", "services"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "create", "delete", "update"},
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"ingresses"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "create", "delete", "update"},
			APIGroups: []string{"networking.x-k8s.io"},
			Resources: []string{"httproutes"},
		},
		{
			Verbs:     []string{"create"},
			APIGroups: []string{"route.openshift.io"},
			Resources: []string{"routes/custom-host"},
		},
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"acme.cert-manager.io"},
			Resources: []string{"challenges/finalizers"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
	},
}

var ControllerChallengesClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-challenges",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-controller-challenges",
	},
}

var ControllerIngressShimClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-ingress-shim",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"create", "update", "delete"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"certificates", "certificaterequests"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"certificates", "certificaterequests", "issuers", "clusterissuers"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"ingresses"},
		},
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"ingresses/finalizers"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"networking.x-k8s.io"},
			Resources: []string{"gateways", "httproutes"},
		},
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"networking.x-k8s.io"},
			Resources: []string{"gateways/finalizers", "httproutes/finalizers"},
		},
		{
			Verbs:     []string{"create", "patch"},
			APIGroups: []string{""},
			Resources: []string{"events"},
		},
	},
}

var ControllerIngressShimClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-controller-ingress-shim",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-controller",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-controller-ingress-shim",
	},
}

var CAInjectorServiceAccount = &corev1.ServiceAccount{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-cainjector",
	},
}

var CAInjectorRole = &rbacv1.Role{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "ibm-cert-manager-cainjector:leaderelection",
		Namespace: "ibm-common-services",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:         []string{"get", "update", "patch"},
			APIGroups:     []string{""},
			Resources:     []string{"configmaps"},
			ResourceNames: []string{"cert-manager-cainjector-leader-election", "cert-manager-cainjector-leader-election-core"},
		},
		{
			Verbs:     []string{"create"},
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
		},
		{
			Verbs:         []string{"get", "update", "patch"},
			APIGroups:     []string{"coordination.k8s.io"},
			Resources:     []string{"leases"},
			ResourceNames: []string{"cert-manager-cainjector-leader-election", "cert-manager-cainjector-leader-election-core"},
		},
		{
			Verbs:     []string{"create"},
			APIGroups: []string{"coordination.k8s.io"},
			Resources: []string{"leases"},
		},
	},
}

var CAInjectorRoleBinding = &rbacv1.RoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "ibm-cert-manager-cainjector:leaderelection",
		Namespace: "ibm-common-services",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "ibm-cert-manager-cainjector",
			Namespace: DeployNamespace,
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "ibm-cert-manager-cainjector:leaderelection",
	},
}

var CAInjectorClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-cainjector",
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"cert-manager.io"},
			Resources: []string{"certificates"},
		},
		{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
		{
			Verbs:     []string{"get", "create", "update", "patch"},
			APIGroups: []string{""},
			Resources: []string{"events"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "update"},
			APIGroups: []string{"admissionregistration.k8s.io"},
			Resources: []string{"validatingwebhookconfigurations", "mutatingwebhookconfigurations"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "update"},
			APIGroups: []string{"apiregistration.k8s.io"},
			Resources: []string{"apiservices"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "update"},
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"customresourcedefinitions"},
		},
		{
			Verbs:     []string{"get", "list", "watch", "update"},
			APIGroups: []string{"auditregistration.k8s.io"},
			Resources: []string{"auditsinks"},
		},
	},
}

var CAInjectorClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "ibm-cert-manager-cainjector",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-cainjector",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "ibm-cert-manager-cainjector",
	},
}

// WebhookRoleBinding is the rolebinding used for the cert-manager-webhook's ability to read the extension-apiserver-authentication
var WebhookRoleBinding = &rbacv1.RoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name:      CertManagerWebhookName,
		Namespace: "kube-system",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			APIGroup:  "",
			Name:      ServiceAccount,
			Namespace: DeployNamespace,
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "extension-apiserver-authentication-reader",
	},
}
