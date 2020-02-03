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

	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-cert-manager-operator/pkg/resources"
	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsAPIv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_certmanager")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new CertManager Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	apiextclient, _ := apiextensionclientset.NewForConfig(mgr.GetConfig())
	kubeclient, _ := kubernetes.NewForConfig(mgr.GetConfig())
	return &ReconcileCertManager{client: mgr.GetClient(), kubeclient: kubeclient, apiextclient: apiextclient, scheme: mgr.GetScheme(), recorder: mgr.GetEventRecorderFor("ibm-cert-manager-operator")}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("certmanager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CertManager
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.CertManager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployments and requeue the owner CertManager
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
	})
	if err != nil {
		return err
	}

	// Watch changes to custom resource defintions that are owned by this operator - in case of deletion or changes
	err = c.Watch(&source.Kind{Type: &apiextensionsAPIv1beta1.CustomResourceDefinition{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Name:      a.Meta.GetLabels()["instance-name"],
					Namespace: a.Meta.GetLabels()["instance-namespace"],
				}},
			}
		}),
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileCertManager implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCertManager{}

// ReconcileCertManager reconciles a CertManager object
type ReconcileCertManager struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client       client.Client
	kubeclient   kubernetes.Interface
	apiextclient apiextensionclientset.Interface
	scheme       *runtime.Scheme
	recorder     record.EventRecorder
}

// Reconcile reads that state of the cluster for a CertManager object and makes changes based on the state read
// and what is in the CertManager.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCertManager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CertManager")
	// Fetch the CertManager instance
	instance := &operatorv1alpha1.CertManager{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.V(2).Info("CR instance not found, don't requeue")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	finalizerName := "certmanager.operators.ibm.com"
	// Determine if the certmanager crd is going to be deleted
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object not being deleted, but add our finalizer so we know to remove this object later when it is going to be deleted
		if !containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				log.Error(err, "Error adding the finalizer to the CR")
				return reconcile.Result{}, err
			}
		}
	} else {
		// Object scheduled to be deleted
		if containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			if err := r.deleteExternalResources(instance); err != nil {
				log.Error(err, "Error deleting resources created by this operator")

				return reconcile.Result{}, err
			}

			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				log.Error(err, "Error updating the CR to remove the finalizer")
				return reconcile.Result{}, err
			}

		}
		return reconcile.Result{}, err
	}

	// Check Prerequisites
	if err := r.PreReqs(instance); err != nil {
		log.Error(err, "One or more prerequisites not met, requeueing")
		r.updateStatus(instance, err.Error(), corev1.EventTypeWarning, "PrereqsFailed")
		return reconcile.Result{Requeue: true}, nil
	}
	r.updateStatus(instance, "All prerequisites for deploying cert-manager service found", corev1.EventTypeNormal, "PrereqsMet")

	// Check Deployment itself
	if err := r.deployments(instance); err != nil {
		log.Error(err, "Error with deploying cert-manager, requeueing")
		r.updateStatus(instance, err.Error(), corev1.EventTypeWarning, "Failed")

		return reconcile.Result{Requeue: true}, nil
	}
	r.updateStatus(instance, "Deployed cert-manager successfully", corev1.EventTypeNormal, "Deployed")

	return reconcile.Result{}, nil
}

func (r *ReconcileCertManager) PreReqs(instance *operatorv1alpha1.CertManager) error {
	if err := checkCrds(r.apiextclient.ApiextensionsV1beta1().CustomResourceDefinitions(), instance.Name, instance.Namespace); err != nil {
		log.V(2).Info("Checking CRDs failed")
		return err
	}
	if err := checkNamespace(r.kubeclient.CoreV1().Namespaces()); err != nil {
		log.V(2).Info("Checking namespace failed")
		return err
	}
	if err := checkRbac(instance, r.scheme, r.client); err != nil {
		log.V(2).Info("Checking RBAC failed")
		return err
	}
	return nil
}

func (r *ReconcileCertManager) deployments(instance *operatorv1alpha1.CertManager) error {
	if err := certManagerDeploy(instance, r.client, r.kubeclient, r.scheme); err != nil {
		return err
	}

	if (&instance.Spec.Webhook != nil && instance.Spec.Webhook) || &instance.Spec.Webhook == nil {
		// Deploy webhook and cainjector
		if err := cainjectorDeploy(instance, r.client, r.kubeclient, r.scheme); err != nil {
			return err
		}
		if err := webhookDeploy(instance, r.client, r.kubeclient, r.scheme); err != nil {
			return err
		}
	} else if &instance.Spec.Webhook != nil && !instance.Spec.Webhook {
		// Specified to not deploy the webhook, remove them if they exist
		webhook := removeDeploy(r.kubeclient, res.CertManagerWebhookName, res.DeployNamespace)
		cainjector := removeDeploy(r.kubeclient, res.CertManagerCainjectorName, res.DeployNamespace)
		if !errors.IsNotFound(webhook) {
			log.Error(webhook, "error removing webhook")
			return webhook
		}
		if !errors.IsNotFound(cainjector) {
			log.Error(cainjector, "error removing webhook")
			return cainjector
		}
	}
	return nil
}

// Removes some of the resources created by this controller for the CR including
// The clusterrolebinding, clusterrole, serviceaccount, and the cert-manager deployment
func (r *ReconcileCertManager) deleteExternalResources(instance *operatorv1alpha1.CertManager) error {
	// Remove RBAC
	if err := removeRbac(r.client); err != nil {
		return err
	}
	// Remove cainjector and webhook
	if err := removeDeploy(r.kubeclient, res.CertManagerWebhookName, res.DeployNamespace); err != nil {
		return err
	}
	if err := removeDeploy(r.kubeclient, res.CertManagerCainjectorName, res.DeployNamespace); err != nil {
		return err
	}
	// Remove the cert-manager-controller deployment
	if err := removeDeploy(r.kubeclient, res.CertManagerControllerName, res.DeployNamespace); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileCertManager) updateStatus(instance *operatorv1alpha1.CertManager, message, event, reason string) {
	r.recorder.Event(instance, event, reason, message)
}
