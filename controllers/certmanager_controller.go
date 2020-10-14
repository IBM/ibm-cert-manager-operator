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

package controllers

import (
	"context"
	"reflect"

	admRegv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsAPIv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	apiRegv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/IBM/ibm-cert-manager-operator/api/v1alpha1"
	res "github.com/IBM/ibm-cert-manager-operator/resources"
)

var log = logf.Log.WithName("controller_certmanager")

// CertManagerReconciler reconciles a CertManager object
type CertManagerReconciler struct {
	Client       client.Client
	Kubeclient   kubernetes.Interface
	Apiextclient apiextensionclientset.Interface
	Scheme       *runtime.Scheme
	Recorder     record.EventRecorder
	Ns           string
}

// Reconcile reads that state of the cluster for a CertManager object and makes changes based on the state readand what is in the CertManager.Spec
// Note: The Controller will requeue the Request to be processed again if the returned error is non-nil or Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *CertManagerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	// _ = r.Log.WithValues("certmanager", req.NamespacedName)

	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling CertManager")

	// Fetch the CertManager instance
	instance := &operatorv1alpha1.CertManager{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.V(2).Info("CR instance not found, don't requeue")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if req.Name != "default" {
		msg := "Only one CR named default is allowed"
		log.Info(msg, "request name", req.Name)
		r.updateEvent(instance, msg, corev1.EventTypeWarning, "Not Allowed")
		return ctrl.Result{}, nil
	}

	//Check RHACM
	rhacmErr := checkRhacm(r.Client)
	if rhacmErr == nil {
		// multiclusterhub found, this means RHACM exists
		// Return and don't requeue
		log.Info("RHACM exists")
		r.updateStatus(instance, "IBM Cloud Platform Common Services cert-manager not installed. Red Hat Advanced Cluster Management for Kubernetes cert-manager is already installed and is in use by Common Services")
		return ctrl.Result{}, nil
	}
	log.Info("RHACM does not exist: " + rhacmErr.Error())

	finalizerName := "certmanager.operators.ibm.com"
	// Determine if the certmanager crd is going to be deleted
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object not being deleted, but add our finalizer so we know to remove this object later when it is going to be deleted
		if !containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.Client.Update(context.Background(), instance); err != nil {
				log.Error(err, "Error adding the finalizer to the CR")
				return ctrl.Result{}, err
			}
		}
	} else {
		// Object scheduled to be deleted
		if containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.Client.Update(context.Background(), instance); err != nil {
				log.Error(err, "Error updating the CR to remove the finalizer")
				return ctrl.Result{}, err
			}

		}
		return ctrl.Result{}, err
	}

	log.Info("The namespace", "ns", r.Ns)
	r.updateEvent(instance, "Instance found", corev1.EventTypeNormal, "Initializing")

	// Check Prerequisites
	if err := r.PreReqs(instance); err != nil {
		log.Error(err, "One or more prerequisites not met, requeueing")
		r.updateStatus(instance, "Error deploying cert-manager, prereqs not met")
		r.updateEvent(instance, err.Error(), corev1.EventTypeWarning, "PrereqsFailed")
		return ctrl.Result{Requeue: true}, nil
	}
	r.updateEvent(instance, "All prerequisites for deploying cert-manager service found", corev1.EventTypeNormal, "PrereqsMet")

	// Check Deployment itself
	if err := r.deployments(instance); err != nil {
		log.Error(err, "Error with deploying cert-manager, requeueing")
		r.updateEvent(instance, err.Error(), corev1.EventTypeWarning, "Failed")
		r.updateStatus(instance, "Error deploying cert-manager")
		return ctrl.Result{Requeue: true}, nil
	}
	r.updateEvent(instance, "Deployed cert-manager successfully", corev1.EventTypeNormal, "Deployed")
	r.updateStatus(instance, "Successfully deployed cert-manager")

	return ctrl.Result{}, nil
}

func (r *CertManagerReconciler) PreReqs(instance *operatorv1alpha1.CertManager) error {
	if err := checkCrds(instance, r.Scheme, r.Apiextclient.ApiextensionsV1beta1().CustomResourceDefinitions(), instance.Name, instance.Namespace); err != nil {
		log.V(2).Info("Checking CRDs failed")
		return err
	}
	if err := checkRbac(instance, r.Scheme, r.Client, r.Ns); err != nil {
		log.V(2).Info("Checking RBAC failed")
		return err
	}
	return nil
}

func (r *CertManagerReconciler) deployments(instance *operatorv1alpha1.CertManager) error {
	if err := certManagerDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.Ns); err != nil {
		return err
	}

	if err := configmapWatcherDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.Ns); err != nil {
		return err
	}

	if instance.Spec.Webhook {
		// Check webhook prerequisites
		if err := webhookPrereqs(instance, r.Scheme, r.Client, r.Ns); err != nil {
			return err
		}
		// Deploy webhook and cainjector
		if err := cainjectorDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.Ns); err != nil {
			return err
		}
		if err := webhookDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.Ns); err != nil {
			return err
		}
	} else {
		// Specified to not deploy the webhook, remove them if they exist
		webhook := removeDeploy(r.Kubeclient, res.CertManagerWebhookName, res.DeployNamespace)
		cainjector := removeDeploy(r.Kubeclient, res.CertManagerCainjectorName, res.DeployNamespace)
		if !errors.IsNotFound(webhook) {
			log.Error(webhook, "error removing webhook")
			return webhook
		}
		if !errors.IsNotFound(cainjector) {
			log.Error(cainjector, "error removing webhook")
			return cainjector
		}
		// Remove webhook prerequisites
		if err := removeWebhookPrereqs(r.Client, r.Ns); err != nil {
			return err
		}
	}
	return nil
}

func (r *CertManagerReconciler) updateEvent(instance *operatorv1alpha1.CertManager, message, event, reason string) {
	r.Recorder.Event(instance, event, reason, message)
}

func (r *CertManagerReconciler) updateStatus(instance *operatorv1alpha1.CertManager, message string) {
	if !reflect.DeepEqual(instance.Status.OverallStatus, message) {
		instance.Status.OverallStatus = message
		if err := r.Client.Status().Update(context.TODO(), instance); err != nil {
			log.Error(err, "Error updating instance status")
		}
	}
}

// SetupWithManager adds CertManager controller to the manager
func (r *CertManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.CertManager{}).
		Owns(&appsv1.Deployment{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&apiextensionsAPIv1beta1.CustomResourceDefinition{}).
		Owns(&admRegv1beta1.MutatingWebhookConfiguration{}).
		Owns(&admRegv1beta1.ValidatingWebhookConfiguration{}).
		Owns(&apiRegv1.APIService{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
