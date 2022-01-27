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
	"fmt"
	"reflect"

	certmgr "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1alpha1"
	"golang.org/x/mod/semver"

	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-cert-manager-operator/pkg/resources"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	admRegv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsAPIv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metaerrors "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	apiRegv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	ns, _ := k8sutil.GetWatchNamespace()

	if ns == "" {
		ns = res.DeployNamespace
	}

	return &ReconcileCertManager{
		client:       mgr.GetClient(),
		kubeclient:   kubeclient,
		apiextclient: apiextclient,
		scheme:       mgr.GetScheme(),
		recorder:     mgr.GetEventRecorderFor("ibm-cert-manager-operator"),
		ns:           ns,
	}
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

	// Watch for changes to secondary resource ClusterRoles and requeue the owner CertManager
	err = c.Watch(&source.Kind{Type: &rbacv1.ClusterRole{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ClusterRoleBindings and requeue the owner CertManager
	err = c.Watch(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ServiceAccounts and requeue the owner CertManager
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
	})
	if err != nil {
		return err
	}

	// Watch changes to custom resource defintions that are owned by this operator - in case of deletion or changes
	err = c.Watch(&source.Kind{Type: &apiextensionsAPIv1.CustomResourceDefinition{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
	})
	if err != nil {
		return err
	}

	// Watch changes to mutating webhook configuration that are owned by this operator - in case of deletion or changes
	err = c.Watch(&source.Kind{Type: &admRegv1.MutatingWebhookConfiguration{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
	})
	if err != nil {
		return err
	}
	// Watch changes to validating webhook configuration that are owned by this operator - in case of deletion or changes
	err = c.Watch(&source.Kind{Type: &admRegv1.ValidatingWebhookConfiguration{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
	})
	if err != nil {
		return err
	}
	// Watch changes to apiservice that are owned by this operator - in case of deletion or changes
	err = c.Watch(&source.Kind{Type: &apiRegv1.APIService{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
	})
	if err != nil {
		return err
	}
	// Watch changes to service that are owned by this operator - in case of deletion or changes
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.CertManager{},
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
	ns           string
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

	if request.Name != "default" {
		msg := "Only one CR named default is allowed"
		log.Info(msg, "request name", request.Name)
		r.updateEvent(instance, msg, corev1.EventTypeWarning, "Not Allowed")
		return reconcile.Result{}, nil
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
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.client.Update(context.Background(), instance); err != nil {
				log.Error(err, "Error updating the CR to remove the finalizer")
				return reconcile.Result{}, err
			}

		}
		return reconcile.Result{}, err
	}

	log.Info("The namespace", "ns", r.ns)
	r.updateEvent(instance, "Instance found", corev1.EventTypeNormal, "Initializing")

	//Check RHACM
	rhacmVersion, rhacmNamespace, rhacmErr := CheckRhacm(r.client)
	if rhacmErr != nil {
		// missing RHACM CR or CRD means RHACM does not exist
		if errors.IsNotFound(rhacmErr) || metaerrors.IsNoMatchError(rhacmErr) {
			log.Error(rhacmErr, "Could not find RHACM")
		} else {
			return reconcile.Result{}, rhacmErr
		}
	}
	if rhacmVersion != "" {
		rhacmVersion = "v" + rhacmVersion
		deployOperand := semver.Compare(rhacmVersion, "v2.3")
		log.Info("Detected RHACM is deployed")
		log.Info("RHACM version: " + rhacmVersion)
		log.Info("RHACM namespace: " + rhacmNamespace)

		if deployOperand < 0 {
			log.Info("RHACM version is less than 2.3, so not deploying operand")
			// multiclusterhub found, this means RHACM exists

			// create a secretshare CR to copy clusterissuer secret to the rhacm issuer ns
			rhacmClusterIssuerNamespace := rhacmNamespace + "-issuer"

			log.Info("RHACM exists. Copying " + res.CSCASecretName + " to namespace " + rhacmClusterIssuerNamespace)
			err := copySecret(r.client, res.CSCASecretName, res.DeployNamespace, rhacmClusterIssuerNamespace, res.RhacmSecretShareCRName)
			if err != nil {
				log.Error(err, "Error creating "+res.RhacmSecretShareCRName)
				return reconcile.Result{}, err
			}

			// Return and don't requeue
			r.updateStatus(instance, "IBM Cloud Platform Common Services cert-manager not installed. Red Hat Advanced Cluster Management for Kubernetes cert-manager is already installed and is in use by Common Services")
			return reconcile.Result{}, nil
		}
	}

	log.Info("RHACM does not exist")

	// Check Prerequisites
	if err := r.PreReqs(instance); err != nil {
		log.Error(err, "One or more prerequisites not met, requeueing")
		r.updateStatus(instance, "Error deploying cert-manager, prereqs not met")
		r.updateEvent(instance, err.Error(), corev1.EventTypeWarning, "PrereqsFailed")
		return reconcile.Result{Requeue: true}, nil
	}
	r.updateEvent(instance, "All prerequisites for deploying cert-manager service found", corev1.EventTypeNormal, "PrereqsMet")

	// Check Deployment itself
	if err := r.deployments(instance); err != nil {
		log.Error(err, "Error with deploying cert-manager, requeueing")
		r.updateEvent(instance, err.Error(), corev1.EventTypeWarning, "Failed")
		r.updateStatus(instance, "Error deploying cert-manager")
		return reconcile.Result{Requeue: true}, nil
	}

	r.updateEvent(instance, "Deployed cert-manager successfully", corev1.EventTypeNormal, "Deployed")
	r.updateStatus(instance, "Successfully deployed cert-manager")

	return reconcile.Result{}, nil
}

func (r *ReconcileCertManager) PreReqs(instance *operatorv1alpha1.CertManager) error {
	if err := checkRbac(instance, r.scheme, r.client, r.ns); err != nil {
		log.V(2).Info("Checking RBAC failed")
		return err
	}
	return nil
}

func (r *ReconcileCertManager) deployments(instance *operatorv1alpha1.CertManager) error {
	if err := certManagerDeploy(instance, r.client, r.kubeclient, r.scheme, r.ns); err != nil {
		return err
	}

	if err := configmapWatcherDeploy(instance, r.client, r.kubeclient, r.scheme, r.ns); err != nil {
		return err
	}

	if instance.Spec.Webhook {
		// Check webhook prerequisites
		if err := webhookPrereqs(instance, r.scheme, r.client, r.ns); err != nil {
			return err
		}
		// Deploy webhook and cainjector
		if err := cainjectorDeploy(instance, r.client, r.kubeclient, r.scheme, r.ns); err != nil {
			return err
		}
		if err := webhookDeploy(instance, r.client, r.kubeclient, r.scheme, r.ns); err != nil {
			return err
		}
	} else {
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
		// Remove webhook prerequisites
		if err := removeWebhookPrereqs(r.client, r.ns); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileCertManager) updateEvent(instance *operatorv1alpha1.CertManager, message, event, reason string) {
	r.recorder.Event(instance, event, reason, message)
}

func (r *ReconcileCertManager) updateStatus(instance *operatorv1alpha1.CertManager, message string) {
	if !reflect.DeepEqual(instance.Status.OverallStatus, message) {
		instance.Status.OverallStatus = message
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			log.Error(err, "Error updating instance status")
		}
	}
}

// createIssuer creates CS CA Issuer
func (r *ReconcileCertManager) createIssuer(instance *operatorv1alpha1.CertManager, issuer *certmgr.Issuer) error {

	// Set CertManager instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, issuer, r.scheme); err != nil {
		return err
	}

	// Create the issuer
	err := r.client.Create(context.TODO(), issuer)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("could not create resource: %v", err)
	}

	return nil
}
