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

package certmanager

import (
	"context"

	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-cert-manager-operator/pkg/resources"

	admRegv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	apiRegv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func webhookPrereqs(instance *operatorv1alpha1.CertManager, scheme *runtime.Scheme, client client.Client, ns string) error {
	if err := createRoleBinding(instance, scheme, client); err != nil {
		return err
	}
	if err := service(instance, scheme, client, ns); err != nil {
		return err
	}
	if err := apiService(instance, scheme, client, ns); err != nil {
		return err
	}
	if err := webhooks(instance, scheme, client); err != nil {
		return err
	}
	return nil
}

func removeWebhookPrereqs(client client.Client, ns string) error {
	if err := removeSvc(client, ns); err != nil {
		return err
	}
	if err := removeAPIService(client); err != nil {
		return err
	}
	if err := removeWebhooks(client); err != nil {
		return err
	}
	if err := removeRoleBinding(client); err != nil {
		return err
	}
	return nil
}

func apiService(instance *operatorv1alpha1.CertManager, scheme *runtime.Scheme, client client.Client, ns string) error {
	apiSvc := &apiRegv1.APIService{}
	err := client.Get(context.Background(), types.NamespacedName{Name: res.APISvcName, Namespace: ""}, apiSvc)
	if err != nil && apiErrors.IsNotFound(err) {
		// Create the apiservice spec
		res.APIService.ResourceVersion = ""
		var servingSecret = ns + "/" + res.WebhookServingSecret
		res.APIService.Annotations = map[string]string{"certmanager.k8s.io/inject-ca-from-secret": servingSecret}
		res.APIService.Spec.Service.Namespace = ns
		if err := controllerutil.SetControllerReference(instance, res.APIService, scheme); err != nil {
			log.Error(err, "Error setting controller reference on api service")
		}
		err := client.Create(context.Background(), res.APIService)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeAPIService(client client.Client) error {
	apiSvc := &apiRegv1.APIService{}
	err := client.Get(context.Background(), types.NamespacedName{Name: res.APISvcName, Namespace: ""}, apiSvc)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
	} else {
		if err := client.Delete(context.Background(), apiSvc); err != nil {
			return err
		}
	}
	return nil
}

func webhooks(instance *operatorv1alpha1.CertManager, scheme *runtime.Scheme, client client.Client) error {
	mutating := &admRegv1beta1.MutatingWebhookConfiguration{}
	err := client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: ""}, mutating)
	if err != nil && apiErrors.IsNotFound(err) {
		// Create the mutating webhook spec
		res.MutatingWebhook.ResourceVersion = ""
		if err := controllerutil.SetControllerReference(instance, res.MutatingWebhook, scheme); err != nil {
			log.Error(err, "Error setting controller reference on mutating webhook")
		}
		err := client.Create(context.Background(), res.MutatingWebhook)
		if err != nil {
			return err
		}
	}

	validating := &admRegv1beta1.ValidatingWebhookConfiguration{}
	err = client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: ""}, validating)
	if err != nil && apiErrors.IsNotFound(err) {
		// Create the validating webhook spec
		res.ValidatingWebhook.ResourceVersion = ""
		if err := controllerutil.SetControllerReference(instance, res.ValidatingWebhook, scheme); err != nil {
			log.Error(err, "Error setting controller reference on validating webhook")
		}
		err := client.Create(context.Background(), res.ValidatingWebhook)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeWebhooks(client client.Client) error {
	mutating := &admRegv1beta1.MutatingWebhookConfiguration{}
	err := client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: ""}, mutating)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
	} else {
		if err := client.Delete(context.Background(), mutating); err != nil {
			return err
		}
	}

	validating := &admRegv1beta1.ValidatingWebhookConfiguration{}
	err = client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: ""}, validating)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
	} else {
		if err := client.Delete(context.Background(), validating); err != nil {
			return err
		}
	}
	return nil
}

func service(instance *operatorv1alpha1.CertManager, scheme *runtime.Scheme, client client.Client, ns string) error {
	svc := &corev1.Service{}
	err := client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: ns}, svc)
	if err != nil && apiErrors.IsNotFound(err) {
		// Create the webhook service spec
		res.WebhookSvc.ResourceVersion = ""
		res.WebhookSvc.Spec.ClusterIP = ""
		res.WebhookSvc.Namespace = ns
		if err := controllerutil.SetControllerReference(instance, res.WebhookSvc, scheme); err != nil {
			log.Error(err, "Error setting controller reference on webhook's service")
		}
		err := client.Create(context.Background(), res.WebhookSvc)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeSvc(client client.Client, ns string) error {
	svc := &corev1.Service{}
	err := client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: ns}, svc)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
	} else {
		if err := client.Delete(context.Background(), svc); err != nil {
			return err
		}
	}
	return nil
}

func createRoleBinding(instance *operatorv1alpha1.CertManager, scheme *runtime.Scheme, client client.Client) error {
	log.V(2).Info("Creating role binding")
	roleBinding := &rbacv1.RoleBinding{}
	err := client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: "kube-system"}, roleBinding)
	if err != nil && apiErrors.IsNotFound(err) {
		res.WebhookRoleBinding.ResourceVersion = ""
		if err := controllerutil.SetControllerReference(instance, res.WebhookRoleBinding, scheme); err != nil {
			log.Error(err, "Error setting controller reference on rolebinding")
		}
		err := client.Create(context.Background(), res.WebhookRoleBinding)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeRoleBinding(client client.Client) error {
	roleBinding := &rbacv1.RoleBinding{}
	err := client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: "kube-system"}, roleBinding)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
	} else {
		if err := client.Delete(context.Background(), roleBinding); err != nil {
			return err
		}
	}
	return nil
}
