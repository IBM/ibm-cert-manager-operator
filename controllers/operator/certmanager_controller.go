/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package operator

import (
	"context"
	"reflect"

	res "github.com/ibm/ibm-cert-manager-operator/controllers/resources"
	"golang.org/x/mod/semver"
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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/operator/v1alpha1"
)

var logd = log.Log.WithName("controller_certmanager")

// CertManagerReconciler reconciles a CertManager object
type CertManagerReconciler struct {
	Client       client.Client
	Kubeclient   kubernetes.Interface
	APIextclient apiextensionclientset.Interface
	Scheme       *runtime.Scheme
	Recorder     record.EventRecorder
	NS           string
}

//+kubebuilder:rbac:groups=operator.ibm.com,resources=certmanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.ibm.com,resources=certmanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.ibm.com,resources=certmanagers/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings;clusterroles;rolebindings;roles,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apiregistration.k8s.io",resources=apiservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups="acme.cert-manager.io",resources=challenges;orders,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="acme.cert-manager.io",resources=orders/finalizers;orders/status;challenges/finalizers;challenges/status,verbs=update

//+kubebuilder:rbac:groups="cert-manager.io",resources=certificaterequests,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="cert-manager.io",resources=certificaterequests/finalizers,verbs=update
//+kubebuilder:rbac:groups="cert-manager.io",resources=certificaterequests/status,verbs=update
//+kubebuilder:rbac:groups="cert-manager.io",resources=signers,verbs=approve
//+kubebuilder:rbac:groups="cert-manager.io",resources=clusterissuers,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="cert-manager.io",resources=clusterissuers/status,verbs=update

//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch;update
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/status,verbs=update
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,verbs=sign

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;create;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete

//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses;httproutes,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses/finalizers,verbs=update

//+kubebuilder:rbac:groups="networking.x-k8s.io",resources=httproutes,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups="networking.x-k8s.io",resources=gateways,verbs=get;list;watch
//+kubebuilder:rbac:groups="networking.x-k8s.io",resources=gateways/finalizers;httproutes/finalizers,verbs=update

//+kubebuilder:rbac:groups="route.openshift.io",resources=routes/custom-host,verbs=create

//+kubebuilder:rbac:groups="auditregistration.k8s.io",resources=auditsinks,verbs=get;list;watch;update

//+kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=create

//+kubebuilder:rbac:groups="operator.open-cluster-management.io",resources=multiclusterhubs,verbs=get;list;watch
//+kubebuilder:rbac:groups="ibmcpcs.ibm.com",resources=secretshares,verbs=create;get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CertManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	reqLogger := logd.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling CertManager")
	// Fetch the CertManager instance
	instance := &operatorv1alpha1.CertManager{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logd.V(2).Info("CR instance not found, don't requeue")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if req.Name != "default" {
		msg := "Only one CR named default is allowed"
		logd.Info(msg, "request name", req.Name)
		r.updateEvent(instance, msg, corev1.EventTypeWarning, "Not Allowed")
		return ctrl.Result{}, nil
	}

	finalizerName := "certmanager.operators.ibm.com"
	// Determine if the certmanager crd is going to be deleted
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object not being deleted, but add our finalizer so we know to remove this object later when it is going to be deleted
		if !containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.Client.Update(context.Background(), instance); err != nil {
				logd.Error(err, "Error adding the finalizer to the CR")
				return ctrl.Result{}, err
			}
		}
	} else {
		// Object scheduled to be deleted
		if containsString(instance.ObjectMeta.Finalizers, finalizerName) {
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizerName)
			if err := r.Client.Update(context.Background(), instance); err != nil {
				logd.Error(err, "Error updating the CR to remove the finalizer")
				return ctrl.Result{}, err
			}

		}
		return ctrl.Result{}, err
	}

	logd.Info("The namespace", "ns", r.NS)
	r.updateEvent(instance, "Instance found", corev1.EventTypeNormal, "Initializing")

	//Check RHACM
	rhacmVersion, rhacmNamespace, rhacmErr := CheckRhacm(r.Client)
	if rhacmErr != nil {
		// missing RHACM CR or CRD means RHACM does not exist
		if errors.IsNotFound(rhacmErr) || metaerrors.IsNoMatchError(rhacmErr) {
			logd.Error(rhacmErr, "Could not find RHACM")
		} else {
			return ctrl.Result{}, rhacmErr
		}
	}
	if rhacmVersion != "" {
		rhacmVersion = "v" + rhacmVersion
		deployOperand := semver.Compare(rhacmVersion, "v2.3")
		logd.Info("Detected RHACM is deployed")
		logd.Info("RHACM version: " + rhacmVersion)
		logd.Info("RHACM namespace: " + rhacmNamespace)

		if deployOperand < 0 {
			logd.Info("RHACM version is less than 2.3, so not deploying operand")
			// multiclusterhub found, this means RHACM exists

			// create a secretshare CR to copy clusterissuer secret to the rhacm issuer ns
			rhacmClusterIssuerNamespace := rhacmNamespace + "-issuer"

			logd.Info("RHACM exists. Copying " + res.CSCASecretName + " to namespace " + rhacmClusterIssuerNamespace)
			err := copySecret(r.Client, res.CSCASecretName, res.DeployNamespace, rhacmClusterIssuerNamespace, res.RhacmSecretShareCRName)
			if err != nil {
				logd.Error(err, "Error creating "+res.RhacmSecretShareCRName)
				return ctrl.Result{}, err
			}

			// Return and don't requeue
			r.updateStatus(instance, "IBM Cloud Platform Common Services cert-manager not installed. Red Hat Advanced Cluster Management for Kubernetes cert-manager is already installed and is in use by Common Services")
			return ctrl.Result{}, nil
		}
	}

	logd.Info("RHACM does not exist")

	// Check Prerequisites
	if err := r.PreReqs(instance); err != nil {
		logd.Error(err, "One or more prerequisites not met, requeueing")
		r.updateStatus(instance, "Error deploying cert-manager, prereqs not met")
		r.updateEvent(instance, err.Error(), corev1.EventTypeWarning, "PrereqsFailed")
		return ctrl.Result{Requeue: true}, nil
	}
	r.updateEvent(instance, "All prerequisites for deploying cert-manager service found", corev1.EventTypeNormal, "PrereqsMet")

	// Check Deployment itself
	if err := r.deployments(instance); err != nil {
		logd.Error(err, "Error with deploying cert-manager, requeueing")
		r.updateEvent(instance, err.Error(), corev1.EventTypeWarning, "Failed")
		r.updateStatus(instance, "Error deploying cert-manager")
		return ctrl.Result{Requeue: true}, nil
	}

	r.updateEvent(instance, "Deployed cert-manager successfully", corev1.EventTypeNormal, "Deployed")
	r.updateStatus(instance, "Successfully deployed cert-manager")

	return ctrl.Result{}, nil
}

func (r *CertManagerReconciler) updateEvent(instance *operatorv1alpha1.CertManager, message, event, reason string) {
	r.Recorder.Event(instance, event, reason, message)
}

func (r *CertManagerReconciler) updateStatus(instance *operatorv1alpha1.CertManager, message string) {
	if !reflect.DeepEqual(instance.Status.OverallStatus, message) {
		instance.Status.OverallStatus = message
		if err := r.Client.Status().Update(context.TODO(), instance); err != nil {
			logd.Error(err, "Error updating instance status")
		}
	}
}

func (r *CertManagerReconciler) PreReqs(instance *operatorv1alpha1.CertManager) error {
	if err := checkRbac(instance, r.Scheme, r.Client, r.NS); err != nil {
		logd.V(2).Info("Checking RBAC failed")
		return err
	}
	return nil
}

func (r *CertManagerReconciler) deployments(instance *operatorv1alpha1.CertManager) error {
	if err := certManagerDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.NS); err != nil {
		return err
	}

	if err := configmapWatcherDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.NS); err != nil {
		return err
	}

	if instance.Spec.Webhook {
		// Check webhook prerequisites
		if err := webhookPrereqs(instance, r.Scheme, r.Client, r.NS); err != nil {
			return err
		}
		// Deploy webhook and cainjector
		if err := cainjectorDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.NS); err != nil {
			return err
		}
		if err := webhookDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.NS); err != nil {
			return err
		}
	} else {
		// Specified to not deploy the webhook, remove them if they exist
		webhook := removeDeploy(r.Kubeclient, res.CertManagerWebhookName, res.DeployNamespace)
		cainjector := removeDeploy(r.Kubeclient, res.CertManagerCainjectorName, res.DeployNamespace)
		if !errors.IsNotFound(webhook) {
			logd.Error(webhook, "error removing webhook")
			return webhook
		}
		if !errors.IsNotFound(cainjector) {
			logd.Error(cainjector, "error removing webhook")
			return cainjector
		}
		// Remove webhook prerequisites
		if err := removeWebhookPrereqs(r.Client, r.NS); err != nil {
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
