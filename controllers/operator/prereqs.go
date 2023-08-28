//
// Copyright 2022 IBM Corporation
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

package operator

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1 "github.com/ibm/ibm-cert-manager-operator/apis/operator/v1"
	res "github.com/ibm/ibm-cert-manager-operator/controllers/resources"
)

// Check all RBAC is ready for cert-manager
func checkRbac(instance *operatorv1.CertManagerConfig, scheme *runtime.Scheme, client client.Client, ns string) error {
	if rolesError := roles(instance, scheme, client, ns); rolesError != nil {
		return rolesError
	}
	return nil
}

func roles(instance *operatorv1.CertManagerConfig, scheme *runtime.Scheme, client client.Client, ns string) error {

	if clusterRoleErr := createClusterRole(instance, scheme, client); clusterRoleErr != nil {
		return clusterRoleErr
	}
	if roleErr := createRole(instance, scheme, client, ns); roleErr != nil {
		return roleErr
	}
	if clusterRoleBindingErr := createClusterRoleBinding(instance, scheme, client, ns); clusterRoleBindingErr != nil {
		return clusterRoleBindingErr
	}
	if roleBindingErr := createRoleBinding(instance, scheme, client, ns); roleBindingErr != nil {
		return roleBindingErr
	}
	if serviceAccountErr := createServiceAccount(instance, scheme, client, ns); serviceAccountErr != nil {
		return serviceAccountErr
	}
	return nil
}

func createRole(instance *operatorv1.CertManagerConfig, scheme *runtime.Scheme, client client.Client, namespace string) error {
	logd.V(0).Info("Creating roles")
	for _, r := range res.RolesToCreate.Items {
		logd.V(0).Info("Creating role " + r.Name)
		role := &rbacv1.Role{}
		err := client.Get(context.Background(), types.NamespacedName{Name: r.Name, Namespace: namespace}, role)
		if err != nil && apiErrors.IsNotFound(err) {
			r.ResourceVersion = ""
			r.Namespace = namespace
			if err := controllerutil.SetControllerReference(instance, &r, scheme); err != nil {
				logd.Error(err, "Error setting controller reference on role")
			}
			err := client.Create(context.Background(), &r)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			oldRole := role.DeepCopy()
			role.Rules = r.Rules
			if !equality.Semantic.DeepEqual(oldRole, role) {
				err := client.Update(context.Background(), role)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func createClusterRole(instance *operatorv1.CertManagerConfig, scheme *runtime.Scheme, client client.Client) error {
	logd.V(0).Info("Creating cluster roles")
	for _, r := range res.ClusterRolesToCreate.Items {
		logd.V(0).Info("Creating cluster role " + r.Name)
		clusterRole := &rbacv1.ClusterRole{}
		err := client.Get(context.Background(), types.NamespacedName{Name: r.Name, Namespace: ""}, clusterRole)
		if err != nil && apiErrors.IsNotFound(err) {
			r.ResourceVersion = ""

			if err := controllerutil.SetControllerReference(instance, &r, scheme); err != nil {
				logd.Error(err, "Error setting controller reference on clusterrole")
			}
			err := client.Create(context.Background(), &r)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			oldClusterRole := clusterRole.DeepCopy()
			clusterRole.Rules = r.Rules
			if !equality.Semantic.DeepEqual(oldClusterRole, clusterRole) {
				err := client.Update(context.Background(), clusterRole)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func createClusterRoleBinding(instance *operatorv1.CertManagerConfig, scheme *runtime.Scheme, client client.Client, namespace string) error {
	logd.V(0).Info("Creating cluster role binding")
	for _, b := range res.ClusterRoleBindingsToCreate.Items {
		logd.V(0).Info("Creating cluster role binding " + b.Name)
		clusterRoleBinding := &rbacv1.ClusterRoleBinding{}

		err := client.Get(context.Background(), types.NamespacedName{Name: b.Name, Namespace: ""}, clusterRoleBinding)
		if err != nil && apiErrors.IsNotFound(err) {
			b.ResourceVersion = ""
			for i := range b.Subjects {
				b.Subjects[i].Namespace = namespace
			}
			if err := controllerutil.SetControllerReference(instance, &b, scheme); err != nil {
				logd.Error(err, "Error setting controller reference on clusterrolebinding")
			}
			err := client.Create(context.Background(), &b)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			for i := range b.Subjects {
				b.Subjects[i].Namespace = namespace
			}
			oldClusterRoleBinding := clusterRoleBinding.DeepCopy()
			clusterRoleBinding.RoleRef = b.RoleRef
			clusterRoleBinding.Subjects = b.Subjects
			if !equality.Semantic.DeepEqual(oldClusterRoleBinding, clusterRoleBinding) {
				err := client.Update(context.Background(), clusterRoleBinding)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func createRoleBinding(instance *operatorv1.CertManagerConfig, scheme *runtime.Scheme, client client.Client, namespace string) error {
	logd.V(0).Info("Creating role binding")
	for _, b := range res.RoleBindingsToCreate.Items {
		logd.V(0).Info("Creating role binding " + b.Name)
		roleBinding := &rbacv1.RoleBinding{}

		err := client.Get(context.Background(), types.NamespacedName{Name: b.Name, Namespace: namespace}, roleBinding)
		if err != nil && apiErrors.IsNotFound(err) {
			b.ResourceVersion = ""
			b.Namespace = namespace
			for i := range b.Subjects {
				b.Subjects[i].Namespace = namespace
			}
			if err := controllerutil.SetControllerReference(instance, &b, scheme); err != nil {
				logd.Error(err, "Error setting controller reference on rolebinding")
			}
			err := client.Create(context.Background(), &b)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			for i := range b.Subjects {
				b.Subjects[i].Namespace = namespace
			}
			oldRolebinding := roleBinding.DeepCopy()
			roleBinding.RoleRef = b.RoleRef
			roleBinding.Subjects = b.Subjects
			if !equality.Semantic.DeepEqual(oldRolebinding, roleBinding) {
				err := client.Update(context.Background(), roleBinding)
				if err != nil {
					return err
				}
			}
			err := client.Update(context.Background(), roleBinding)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createServiceAccount(instance *operatorv1.CertManagerConfig, scheme *runtime.Scheme, client client.Client, namespace string) error {
	logd.V(0).Info("Creating service account")
	for _, a := range res.ServiceAccountsToCreate.Items {
		logd.V(0).Info("Creating service account" + a.Name)
		a.ResourceVersion = ""
		a.Namespace = namespace
		err := client.Create(context.Background(), &a)
		if err := controllerutil.SetControllerReference(instance, &a, scheme); err != nil {
			logd.Error(err, "Error setting controller reference on service account")
		}
		if err != nil {
			if !apiErrors.IsAlreadyExists(err) {
				logd.V(2).Info("Error creating the service account, but was not an already exists error", "error message", err)
				return err
			}
		}
	}
	return nil
}
