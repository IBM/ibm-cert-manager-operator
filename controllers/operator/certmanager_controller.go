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
	"fmt"
	"reflect"
	"strings"

	operatorv1 "github.com/ibm/ibm-cert-manager-operator/apis/operator/v1"
	res "github.com/ibm/ibm-cert-manager-operator/controllers/resources"
	admRegv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logd = log.Log.WithName("controller_certmanager")

var ControllerAppLabel = map[string]string{
	"app": "ibm-cert-manager-controller",
}

// CertManagerReconciler reconciles a CertManager object
type CertManagerReconciler struct {
	Client       client.Client
	Reader       client.Reader
	Kubeclient   kubernetes.Interface
	APIextclient apiextensionclientset.Interface
	Scheme       *runtime.Scheme
	Recorder     record.EventRecorder
	NS           string
}

//+kubebuilder:rbac:groups=operator.ibm.com,resources=certmanagerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.ibm.com,resources=certmanagerconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.ibm.com,resources=certmanagerconfigs/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings;clusterroles;rolebindings;roles,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apiregistration.k8s.io",resources=apiservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch;update
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/status,verbs=update
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,verbs=sign

//+kubebuilder:rbac:groups="acme.cert-manager.io",resources=challenges;orders,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="acme.cert-manager.io",resources=orders/finalizers;orders/status;challenges/finalizers;challenges/status,verbs=update

//+kubebuilder:rbac:groups="cert-manager.io",resources=certificaterequests,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="cert-manager.io",resources=certificaterequests/finalizers,verbs=update
//+kubebuilder:rbac:groups="cert-manager.io",resources=certificaterequests/status,verbs=update
//+kubebuilder:rbac:groups="cert-manager.io",resources=signers,verbs=approve
//+kubebuilder:rbac:groups="cert-manager.io",resources=clusterissuers,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="cert-manager.io",resources=clusterissuers/status,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets;daemonsets,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates/finalizers,verbs=update
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;create;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list

//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses;httproutes,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses/finalizers,verbs=update

//+kubebuilder:rbac:groups="networking.x-k8s.io",resources=httproutes,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups="networking.x-k8s.io",resources=gateways,verbs=get;list;watch
//+kubebuilder:rbac:groups="networking.x-k8s.io",resources=gateways/finalizers;httproutes/finalizers,verbs=update

//+kubebuilder:rbac:groups="route.openshift.io",resources=routes/custom-host,verbs=create

//+kubebuilder:rbac:groups="auditregistration.k8s.io",resources=auditsinks,verbs=get;list;watch;update

//+kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=create

//+kubebuilder:rbac:groups="ibmcpcs.ibm.com",resources=secretshares,verbs=create;get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CertManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	reqLogger := logd.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling CertManager")
	// Fetch the CertManager instance
	instance := &operatorv1.CertManagerConfig{}
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

	finalizerName := "certmanager.operators.ibm.com"
	// Determine if the certmanager crd is going to be deleted
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
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

	if !instance.Spec.License.Accept {
		logd.Error(nil, "Accept license by changing .spec.license.accept to true in the CertManagerConfig CR. This message will keep showing until then")
	}

	if err := r.updateLabels(ctx); err != nil {
		logd.Error(err, "Error with updating cert-manager labels, requeueing")
		r.updateStatus(instance, "Error updating cert-manager labels")
		r.updateEvent(instance, err.Error(), corev1.EventTypeWarning, "LabelsFailed")
		return ctrl.Result{Requeue: true}, nil
	}

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

	if err := r.updateVersion(instance); err != nil {
		logd.Error(err, "Error updating certmanagerconfig cr")
		r.updateEvent(instance, err.Error(), corev1.EventTypeWarning, "Failed")
		r.updateStatus(instance, "Error updating version")
		return ctrl.Result{Requeue: true}, nil
	}

	r.updateEvent(instance, "Deployed cert-manager successfully", corev1.EventTypeNormal, "Deployed")
	r.updateStatus(instance, "Successfully deployed cert-manager")
	return ctrl.Result{}, nil
}

func (r *CertManagerReconciler) updateEvent(instance *operatorv1.CertManagerConfig, message, event, reason string) {
	r.Recorder.Event(instance, event, reason, message)
}

func (r *CertManagerReconciler) updateStatus(instance *operatorv1.CertManagerConfig, message string) {
	if !reflect.DeepEqual(instance.Status.OverallStatus, message) {
		instance.Status.OverallStatus = message
		if err := r.Client.Status().Update(context.TODO(), instance); err != nil {
			logd.Error(err, "Error updating instance status")
		}
	}
}

func (r *CertManagerReconciler) PreReqs(instance *operatorv1.CertManagerConfig) error {
	if err := checkRbac(instance, r.Scheme, r.Client, r.NS); err != nil {
		logd.V(2).Info("Checking RBAC failed")
		return err
	}
	return nil
}

func (r *CertManagerReconciler) deployments(instance *operatorv1.CertManagerConfig) error {
	if err := certManagerDeploy(instance, r.Client, r.Kubeclient, r.Scheme, r.NS); err != nil {
		return err
	}

	if err := removeDeploy(r.Kubeclient, res.ConfigmapWatcherName, r.NS); err != nil {
		return err
	}

	if instance.Spec.Webhook {
		// Check webhook prerequisites
		if err := webhookPrereqs(instance, r.Scheme, r.Client, r.Reader, r.NS); err != nil {
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

// CreateObject create k8s resource with the unstructured object
func (r *CertManagerReconciler) CreateObject(obj *unstructured.Unstructured) error {
	err := r.Client.Create(context.TODO(), obj)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("could not Create resource: %v", err)
	}
	return nil
}

// DeleteObject delete k8s resource with the unstructured object
func (r *CertManagerReconciler) DeleteObject(obj *unstructured.Unstructured) error {
	err := r.Client.Delete(context.TODO(), obj)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("could not Delete resource: %v", err)
	}
	return nil
}

// UpdateObject update k8s resource with the unstructured object
func (r *CertManagerReconciler) UpdateObject(obj *unstructured.Unstructured) error {
	if err := r.Client.Update(context.TODO(), obj); err != nil {
		return fmt.Errorf("could not update resource: %v", err)
	}
	return nil
}

// Updating resource and add resourceVersion
func (r *CertManagerReconciler) UpdateResourse(obj *unstructured.Unstructured, crd *unstructured.Unstructured, labels map[string]string) error {
	gvk := obj.GetObjectKind().GroupVersionKind()

	klog.Infof("Updating resource with name: %s, namespace: %s, kind: %s, apiversion: %s/%s\n", obj.GetName(), obj.GetNamespace(), gvk.Kind, gvk.Group, gvk.Version)
	resourceVersion := crd.GetResourceVersion()
	obj.SetResourceVersion(resourceVersion)
	obj.SetLabels(labels)
	if err := r.UpdateObject(obj); err != nil {
		return err
	}

	return nil
}

// GetObject get k8s resource with the unstructured object
func (r *CertManagerReconciler) GetObject(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	found := &unstructured.Unstructured{}
	found.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	err := r.Reader.Get(context.TODO(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, found)

	return found, err
}

// create CertManagerConfig CR if not exist or update it if the version is old
func (r *CertManagerReconciler) CreateCertManagerConfigCR() error {
	klog.Infof("Creating CertManagerConfig CR")
	var errMsg error

	objects, err := YamlToObjects([]byte(res.CertManagerConfigCR))
	if err != nil {
		return err
	}

	for _, obj := range objects {
		gvk := obj.GetObjectKind().GroupVersionKind()
		_, err := r.GetObject(obj)
		//this object not exist we need to create it
		if errors.IsNotFound(err) {
			klog.Infof("Creating certmanagerconfig CR with name: %s, kind: %s, apiversion: %s/%s\n", obj.GetName(), gvk.Kind, gvk.Group, gvk.Version)
			if e := r.CreateObject(obj); e != nil {
				errMsg = e
			}
		} else if err == nil {
			klog.Infof("Found certmanagerconfig CR, skip creating")
		} else if err != nil {
			klog.Infof("can't get object:%s: %v", obj.GetName(), err)
			errMsg = err
		}
	}

	return errMsg
}

// update version of CertManagerConfig CR
func (r *CertManagerReconciler) updateVersion(instance *operatorv1.CertManagerConfig) error {
	name := "ibm-cert-manager-operator"
	namespace := r.NS
	deployKey := types.NamespacedName{Name: name, Namespace: namespace}
	deploy := &appsv1.Deployment{}
	if err := r.Reader.Get(context.TODO(), deployKey, deploy); err != nil {
		klog.Errorf("Failed to get deployment %s/%s, %s", namespace, name, err)
		return err
	}

	if csv, ok := deploy.GetLabels()["olm.owner"]; ok {
		csvVersion := strings.SplitN(csv, ".", 2)
		version := strings.Replace(csvVersion[1], "v", "", 1)

		if instance.Spec.Version != version {
			instance.Spec.Version = version
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				logd.Error(err, "Error updating instance version")
				return err
			}
		}
	}

	return nil
}

func (r *CertManagerReconciler) updateLabels(ctx context.Context) error {
	// update LabelMaps with Original LabelMaps
	ClearLabelMap(res.ControllerLabelMap)
	ClearLabelMap(res.CainjectorLabelMap)
	ClearLabelMap(res.WebhookLabelMap)

	for key, val := range res.OriginalControllerLabelMap {
		(res.ControllerLabelMap)[key] = val
		logd.Info("Controller Label Message:", fmt.Sprint(res.WebhookLabelMap))
	}

	for key, val := range res.OriginalCainjectorLabelMap {
		(res.CainjectorLabelMap)[key] = val
		logd.Info("CA Label Message:", fmt.Sprint(res.WebhookLabelMap))
	}

	for key, val := range res.OriginalWebhookLabelMap {
		(res.WebhookLabelMap)[key] = val
		logd.Info("Webhook Label Message:", fmt.Sprint(res.WebhookLabelMap))
	}

	// ADD new label to the Labelmaps
	// list all the resources in the cluster
	instanceList := &operatorv1.CertManagerConfigList{}
	if err := r.Client.List(ctx, instanceList); err != nil {
		return err
	}

	for _, instance := range instanceList.Items {
		labels := instance.Spec.Labels
		for key, val := range labels {
			(res.WebhookLabelMap)[key] = val
			(res.ControllerLabelMap)[key] = val
			(res.CainjectorLabelMap)[key] = val
			logd.Info("Webhook Label Message:", fmt.Sprint(res.WebhookLabelMap))
		}
	}

	return nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *CertManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create certManager CRDs
	if err := r.CreateCertManagerConfigCR(); err != nil {
		klog.Errorf("Fail to create CertManager Instance: %v", err)
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("certmanagerconfig_controller").
		For(&operatorv1.CertManagerConfig{}).
		Owns(&appsv1.Deployment{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&admRegv1.MutatingWebhookConfiguration{}).
		Owns(&admRegv1.ValidatingWebhookConfiguration{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
