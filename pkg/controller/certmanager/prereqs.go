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

package certmanager

import (
	"context"
	"fmt"
	"strings"

	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-cert-manager-operator/pkg/resources"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionclientsetv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Check all RBAC is ready for cert-manager
func checkRbac(instance *operatorv1alpha1.CertManager, scheme *runtime.Scheme, client client.Client) error {
	if imagePullSecretError := imagePullSecret(scheme, client, instance); imagePullSecretError != nil {
		return imagePullSecretError
	}
	if rolesError := roles(client); rolesError != nil {
		return rolesError
	}
	return nil
}

// Check that the image pull secret exists in the deploy namespace (cert-manager)
// returns nil if it does, an error otherwise
// We never create the image pull secret since it contains credentials. We only copy it or use the one provided.
func imagePullSecret(scheme *runtime.Scheme, client client.Client, instance *operatorv1alpha1.CertManager) error {
	pullSecret := &corev1.Secret{}
	copyPullSecret := &corev1.Secret{}

	name := res.ImagePullSecret
	namespace := res.DeployNamespace

	pullSecretExists := true
	copyPullSecretExists := false

	if instance.Spec.PullSecret.Name != "" {
		name = instance.Spec.PullSecret.Name
	}

	err := client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, pullSecret)
	if err != nil && apiErrors.IsNotFound(err) { // Pull secret does not already exist in namespace
		pullSecretExists = false
	}

	// Get secret from the namespace in the spec and copy it over to the cert-manager namespace
	if instance.Spec.PullSecret.Namespace != "" {
		err := client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: instance.Spec.PullSecret.Namespace}, copyPullSecret)
		if err != nil && apiErrors.IsNotFound(err) {
			log.V(2).Info("Image pull secret not found in specified namespace", "pull secret name", name, "pull secret namespace", instance.Spec.PullSecret.Namespace)
		} else {
			copyPullSecretExists = true
		}
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data:       copyPullSecret.Data,
		StringData: copyPullSecret.StringData,
		Type:       copyPullSecret.Type,
	}

	if pullSecretExists && copyPullSecretExists { // Perform update to existing pull secret
		if err = client.Update(context.Background(), secret); err != nil {
			return err
		}
		log.V(2).Info("Updated image pull secret")
		controllerutil.SetControllerReference(instance, secret, scheme)
	} else if copyPullSecretExists && !pullSecretExists { // Copy it over using create
		if err = client.Create(context.Background(), secret); err != nil {
			return err
		}
		log.V(2).Info("Created image pull secret")
		controllerutil.SetControllerReference(instance, secret, scheme)
	} else if !copyPullSecretExists && !pullSecretExists { // Secret not found at all, throw an error
		errorMsg := apiErrors.NewNotFound(corev1.Resource("secrets"), fmt.Sprintf("The image pull secret %s does not exist in the deploy namespace %s and there was no copy pull secret found in the %s namespace", name, namespace, instance.Spec.PullSecret.Namespace))
		log.Error(errorMsg, "Neither pull secret exist")
		return errorMsg
	}
	// Pull secret exists and there's no copy pull secret
	log.V(2).Info("Pull secret exists")
	return nil
}

func roles(client client.Client) error {
	// Remove any roles that exist already and create them fresh
	if err := removeRoles(client); err != nil {
		return err
	}
	if clusterRoleErr := createClusterRole(client); clusterRoleErr != nil {
		return clusterRoleErr
	}
	if clusterRoleBindingErr := createClusterRoleBinding(client); clusterRoleBindingErr != nil {
		return clusterRoleBindingErr
	}
	if serviceAccountErr := createServiceAccount(client); serviceAccountErr != nil {
		return serviceAccountErr
	}
	return nil
}

func createClusterRole(client client.Client) error {
	log.V(2).Info("Creating cluster role")
	res.DefaultClusterRole.ResourceVersion = ""
	err := client.Create(context.Background(), res.DefaultClusterRole)
	if err != nil {
		return err
	}
	return nil
}

func createClusterRoleBinding(client client.Client) error {
	log.V(2).Info("Creating cluster role binding")
	res.DefaultClusterRoleBinding.ResourceVersion = ""
	err := client.Create(context.Background(), res.DefaultClusterRoleBinding)
	if err != nil {
		return err
	}
	return nil
}

func createServiceAccount(client client.Client) error {
	log.V(2).Info("Creating service account")
	err := client.Create(context.Background(), res.DefaultServiceAccount)
	if err != nil {
		if !apiErrors.IsAlreadyExists(err) {
			return err
		}
		log.V(2).Info("Error creating the service account, but was not an already exists error", "error message", err)
	}
	return nil
}

// Checks to ensure the namespace we're deploying the service in exists
func checkNamespace(client v1.NamespaceInterface) error {
	getOpt := metav1.GetOptions{}

	if _, err := client.Get(res.DeployNamespace, getOpt); err != nil && apiErrors.IsNotFound(err) {
		log.V(1).Info("cert-manager namespace does not exist, creating it", "error", err)
		if _, err = client.Create(res.NamespaceDef); err != nil {
			return err
		}
	}
	log.V(2).Info("cert-manager namespace exists")
	return nil
}

// Checks for the existence of all certmanager CRDs
// Takes action to create them if they do not exist
func checkCrds(client apiextensionclientsetv1beta1.CustomResourceDefinitionInterface, name, namespace string) error {
	var allErrors []string
	listOptions := metav1.ListOptions{LabelSelector: res.ControllerLabels}
	customResourcesList, err := client.List(listOptions)
	if err != nil {
		return err
	}

	existingResources := make(map[string]bool)
	for _, item := range customResourcesList.Items {
		existingResources[item.Name] = false
	}

	// Check that the CRDs we need match the ones we got from the cluster
	for _, item := range res.CRDs {
		crName := item + "." + res.GroupVersion
		if _, ok := existingResources[crName]; !ok { // CRD wasn't found, create it
			log.V(1).Info("Did not find custom resource, creating it now", "resource", item)
			crd := res.CRDMap[item]
			crd.ObjectMeta.Labels["instance-name"] = name
			crd.ObjectMeta.Labels["instance-namespace"] = namespace

			if _, err = client.Create(crd); err != nil {
				allErrors = append(allErrors, err.Error())
			}
		}
	}
	if allErrors != nil {
		return errors.New(strings.Join(allErrors, "\n"))
	}
	log.V(2).Info("Finished checking CRDs, no errors found")
	return nil
}

// Removes the clusterrole and clusterrolebinding created by this operator
func removeRoles(client client.Client) error {
	// Delete the clusterrolebinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}

	if err := client.Get(context.Background(), types.NamespacedName{Name: res.ClusterRoleName, Namespace: ""}, clusterRoleBinding); err != nil && apiErrors.IsNotFound(err) {
		log.V(1).Info("Error getting cluster role binding", "msg", err)
		return nil
	} else if err == nil {
		if err = client.Delete(context.Background(), clusterRoleBinding); err != nil {
			log.V(1).Info("Error deleting cluster role binding", "name", clusterRoleBinding.Name, "error message", err)
			return err
		}
	} else {
		return err
	}
	// Delete the clusterrole
	clusterRole := &rbacv1.ClusterRole{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: res.ClusterRoleName, Namespace: ""}, clusterRole); err != nil && apiErrors.IsNotFound(err) {
		log.V(1).Info("Error getting cluster role", "msg", err)
		return nil
	} else if err == nil {
		if err = client.Delete(context.Background(), clusterRole); err != nil {
			log.V(1).Info("Error deleting cluster role", "name", clusterRole.Name, "error message", err)
			return err
		}
	} else {
		return err
	}
	return nil
}

// Removes all RBAC resources created by this operator
// Includes: image pull secret, clusterrole, clusterrolebinding, and service account
func removeRbac(client client.Client) error {
	// Delete the pull secret
	pullSecret := &corev1.Secret{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: res.ImagePullSecret, Namespace: res.DeployNamespace}, pullSecret); err != nil && apiErrors.IsNotFound(err) {
		log.V(1).Info("Error getting pull secret", "msg", err)
		return nil
	} else if err == nil {
		if err = client.Delete(context.Background(), pullSecret); err != nil {
			log.V(1).Info("Error deleting pull secret", "name", pullSecret.Name, "error message", err)
			return err
		}
	} else {
		return err
	}

	// Delete the clusterrolebinding & clusterrole
	if err := removeRoles(client); err != nil {
		return err
	}
	// Delete the service account - maybe we shouldn't remove this?
	serviceAccount := &corev1.ServiceAccount{}
	if err := client.Get(context.Background(), types.NamespacedName{Name: res.ServiceAccount, Namespace: res.DeployNamespace}, serviceAccount); err != nil && apiErrors.IsNotFound(err) {
		log.V(1).Info("Error getting service account", "msg", err)
		return nil
	} else if err == nil {
		if err = client.Delete(context.Background(), serviceAccount); err != nil {
			log.V(1).Info("Error deleting service account", "name", serviceAccount.Name, "error message", err)
			return err
		}
	} else {
		return err
	}
	return nil
}
