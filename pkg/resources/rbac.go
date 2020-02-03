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
	"k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultServiceAccount = &corev1.ServiceAccount{
	ObjectMeta: metav1.ObjectMeta{
		Name:      ServiceAccount,
		Namespace: DeployNamespace,
	},
}

var DefaultClusterRole = &v1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: ClusterRoleName,
	},
	Rules: []v1.PolicyRule{
		{
			Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			APIGroups: []string{""},
			Resources: []string{"secrets"},
		},
		{
			Verbs:     []string{"*"},
			APIGroups: []string{"certmanager.k8s.io"},
			Resources: []string{"certificates", "issuers", "clusterissuers", "orders", "challenges"},
		},
		{
			Verbs:     []string{"update"},
			APIGroups: []string{"certmanager.k8s.io"},
			Resources: []string{"certificates/status", "certificaterequests/status", "challenges/status", "clusterissuers/status", "issuers/status", "orders/status", "certificates/finalizers", "challenges/finalizers", "ingresses/finalizers", "orders/finalizers"},
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
			Verbs:     []string{"*"},
			APIGroups: []string{"apps"},
			Resources: []string{"deployments", "statefulsets", "daemonsets"},
		},
	},
}

var DefaultClusterRoleBinding = &v1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: ClusterRoleName,
	},
	Subjects: []v1.Subject{
		{
			Kind:      "ServiceAccount",
			APIGroup:  "",
			Name:      ServiceAccount,
			Namespace: DeployNamespace,
		},
	},
	RoleRef: v1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     ClusterRoleName,
	},
}
