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
	Items: []corev1.ServiceAccount{*DefaultServiceAccount, *CAInjectorServiceAccount},
}

var RolesToCreate = &rbacv1.RoleList{
	Items: []rbacv1.Role{*CAInjectorRole},
}

var RoleBindingsToCreate = &rbacv1.RoleBindingList{
	Items: []rbacv1.RoleBinding{*CAInjectorRoleBinding},
}

var ClusterRolesToCreate = &rbacv1.ClusterRoleList{
	Items: []rbacv1.ClusterRole{*DefaultClusterRole, *CAInjectorClusterRole},
}

var ClusterRoleBindingsToCreate = &rbacv1.ClusterRoleBindingList{
	Items: []rbacv1.ClusterRoleBinding{*DefaultClusterRoleBinding, *CAInjectorClusterRoleBinding},
}

// DefaultServiceAccount is the service account used by cert-manager service
var DefaultServiceAccount = &corev1.ServiceAccount{
	ObjectMeta: metav1.ObjectMeta{
		Name: ServiceAccount,
		//		Namespace: DeployNamespace,
	},
}

// DefaultClusterRole is the cluster role used by cert-manager service
var DefaultClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: ClusterRoleName,
	},
	Rules: []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			APIGroups: []string{""},
			Resources: []string{"secrets", "configmaps"},
		},
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"certmanager.k8s.io"},
			Resources: []string{"certificates", "issuers", "clusterissuers", "orders", "challenges"},
		},
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"certmanager.k8s.io"},
			Resources: []string{
				"certificates/status",
				"certificaterequests/status",
				"challenges/status",
				"clusterissuers/status",
				"issuers/status",
				"orders/status",
				"certificates/finalizers",
				"challenges/finalizers",
				"ingresses/finalizers",
				"orders/finalizers",
			},
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
			APIGroups: []string{"extensions"},
			Resources: []string{"ingresses"},
		},
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"apps"},
			Resources: []string{"deployments", "statefulsets", "daemonsets"},
		},
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"customresourcedefinitions"},
		},
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"admission.certmanager.k8s.io"},
			Resources: []string{"certificates", "clusterissuers", "issuers", "certificaterequests"},
		},
		{
			Verbs:         []string{"use"},
			APIGroups:     []string{"security.openshift.io"},
			Resources:     []string{"securitycontextconstraints"},
			ResourceNames: []string{"restricted", "hostnetwork"},
		},
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"admissionregistration.k8s.io"},
			Resources: []string{"mutatingwebhookconfigurations", "validatingwebhookconfigurations"},
		},
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"apiregistration.k8s.io"},
			Resources: []string{"apiservices"},
		},
		{
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			APIGroups: []string{"authorization.k8s.io"},
			Resources: []string{"subjectaccessreviews"},
		},
	},
}

// DefaultClusterRoleBinding the clusterrolebinding used by cert-manager service
var DefaultClusterRoleBinding = &rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: ClusterRoleName,
	},
	Subjects: []rbacv1.Subject{
		{
			Kind:     "ServiceAccount",
			APIGroup: "",
			Name:     ServiceAccount,
			//			Namespace: DeployNamespace,
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     ClusterRoleName,
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
			Kind: "ServiceAccount",
			Name: "ibm-cert-manager-cainjector",
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
