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
	"time"

	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/apis/cert-manager/v1"
	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/operator/v1alpha1"
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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	apiRegv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var logd = log.Log.WithName("controller_certmanager")

var managedbyLabel = map[string]string{
	"app.kubernetes.io/managed-by": "ibm-cert-manager-operator",
}

var old_labels = map[string]string{
	"operators.coreos.com/ibm-cert-manager-operator." + res.DeployNamespace: "",
}

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

	configMapName := "ibm-cpp-config"
	conditionalDeployCM := &corev1.ConfigMap{}
	if err := r.Reader.Get(context.TODO(), types.NamespacedName{
		Name:      configMapName,
		Namespace: res.DeployNamespace,
	}, conditionalDeployCM); err != nil {
		if errors.IsNotFound(err) {
			logd.Info("ibm-cpp-config ConfigMap does not exist, continuing...")
		} else {
			logd.Error(err, "Failed to get ibm-cpp-config ConfigMap, reconciling...")
			return ctrl.Result{}, err
		}
	}

	if v, ok := conditionalDeployCM.Data["deployCSCertManagerOperands"]; ok {
		if v == "false" {
			logd.Info("deployCSCertManagerOperand value in ibm-cpp-configmap is false, so skipping operand installation")
			return ctrl.Result{}, nil
		}
	}

	logd.Info("Starting auto-detection process to search for another cert-manager running on cluster")

	// delete Issuer in case it was not cleaned up properly before
	if err = r.DeleteIssuer(res.Issuer); err != nil {
		logd.Info("Failed to clean up auto-detection resources from previous checks")
		return ctrl.Result{}, err
	}

	foundCommunityCertManager, err := r.FoundCommunityCertManager(res.Issuer)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "failed to call webhook") {
			isBedrockWebhook, err := r.checkValidatingWebhookConfiguration()
			if err != nil {
				if !errors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
				isBedrockWebhook, err = r.checkMutatingWebhookConfiguration()
				if err != nil {
					return ctrl.Result{}, err
				}
			}

			// If the webhook error is from the community, then user should resolve them since not managed by Bedrock
			// If error is coming from Bedrock cert-manager-webhooks, we continue with normal operand installation
			// since this was how the behaviour previously was, and maybe a reconciliation can fix things.
			if !isBedrockWebhook {
				logd.Info("Auto-detection found error with calling cert-manager-webhook, verify your open source cert-manager installation, and then restart this pod")
				return ctrl.Result{}, nil
			}

		} else {
			logd.Error(nil, "Auto-detection found error while checking if another cert-manager installed")
			return ctrl.Result{}, err
		}
	}

	logd.Info("Auto-detection process complete")

	if foundCommunityCertManager {
		logd.Info("Auto-detection found another cert-manager running on cluster, so skipping operand reconcile")
		r.updateEvent(instance, "Found another cert-manager running on cluster, skipping operand deployment", corev1.EventTypeNormal, "Skipped")
		r.updateStatus(instance, "Successfully skipped operand deployment because another cert-manager running on cluster")
		return ctrl.Result{}, nil
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

	if err := removeDeploy(r.Kubeclient, res.ConfigmapWatcherName, r.NS); err != nil {
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

// check if the name and label of ValidatingWebhookConfiguration is belong to ibm
func (r *CertManagerReconciler) checkValidatingWebhookConfiguration() (bool, error) {
	validating := &admRegv1.ValidatingWebhookConfiguration{}
	// check the name of this ValidatingWebhookConfiguration
	err := r.Client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: ""}, validating)
	if err != nil {
		logd.Error(err, "Failed to get ValidatingWebhookConfiguration", "name:", res.CertManagerWebhookName)
		return false, err
	}

	label := validating.GetLabels()["app"]
	if label != "ibm-cert-manager-webhook" {
		return false, nil
	}

	return true, nil
}

func (r *CertManagerReconciler) checkMutatingWebhookConfiguration() (bool, error) {
	webhook := &admRegv1.MutatingWebhookConfiguration{}
	err := r.Client.Get(context.Background(), types.NamespacedName{Name: res.CertManagerWebhookName, Namespace: ""}, webhook)
	if err != nil {
		logd.Error(err, "Failed to get MutatingWebhookConfiguration", "name:", res.CertManagerWebhookName)
		return false, err
	}

	label := webhook.GetLabels()["app"]
	if label != "ibm-cert-manager-controller" {
		return false, nil
	}

	return true, nil
}

// used to check communityCertManager is deployed or not
// try to deploy issuer if it has status, and check cert-manager-controller deployment has ibm label or not
func (r *CertManagerReconciler) FoundCommunityCertManager(v1Issuer certmanagerv1.Issuer) (bool, error) {

	logd.Info("Creating Issuer for auto-detection. If Issuer is reconciled and cannot find Bedrock cert-manager-controller, then another cert-manager is running on the cluster.")
	if err := r.CreateIssuer(v1Issuer); err != nil {
		logd.Info("Checking if error is from webhook")
		return false, err
	}

	hasStatus, err := r.hasStatus()
	if err != nil {
		logd.Error(err, "Failed to check status of auto-detection Issuer", "name:", res.Issuer.Name, "namespace:", res.Issuer.Namespace)
		return false, err
	}

	if err := r.DeleteIssuer(v1Issuer); err != nil {
		return false, err
	}

	// in upgrade scenario, Bedrock cert-manager controller could be running
	// and if it is, then continue with operand creation logic
	if hasStatus {
		if err := r.GetCertManagerControllerDeployment(); err != nil {
			if errors.IsNotFound(err) {
				logd.Info("Auto-detection could not find Bedrock cert-manager-controller")
				return true, nil
			}
			logd.Error(err, "Auto-detection encountered error finding Bedrock cert-manager-controller")
			return false, err
		}

		logd.Info("Auto-detection found Bedrock cert-manager-controller, continuing operand reconcile")
		return false, nil
	}

	logd.Info("Auto-detection did not find any cert-manager-controller found on cluster, continuing operand reconcile")
	return false, nil
}

// get the status of example issuer
func (r *CertManagerReconciler) hasStatus() (bool, error) {
	pollRate := 1 * time.Second
	timeout := 5

	// extend wait time if CRDs are found to handle slow clusters
	crd := &apiextensionsAPIv1.CustomResourceDefinition{}
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: "issuers.cert-manager.io"}, crd); err != nil {
		if !errors.IsNotFound(err) {
			return false, err
		}
		crd = nil
	}

	if crd != nil {
		timeout = timeout * 6
	}

	issuer := &certmanagerv1.Issuer{}
	for i := 0; i < timeout; i++ {
		logd.Info("Polling auto-detection Issuer status", "poll rate:", pollRate, "timeout:", pollRate*time.Duration(timeout))
		if err := r.Client.Get(context.TODO(), types.NamespacedName{Namespace: res.Issuer.Namespace, Name: res.Issuer.Name}, issuer); err != nil {
			// ignore not found errors because k8s might be slow to create the smoke-check-issuer
			if !errors.IsNotFound(err) {
				return false, err
			}
		}
		if issuer.Status.Conditions != nil {
			return true, nil
		}
		time.Sleep(pollRate)
	}

	return false, nil
}

// Creates an Issuer issuer
func (r *CertManagerReconciler) CreateIssuer(i certmanagerv1.Issuer) error {
	err := r.Client.Create(context.TODO(), &i)
	if err != nil && !errors.IsAlreadyExists(err) {
		logd.Info("Failed to create Issuer", "name:", res.Issuer.Name, "namespace:", res.Issuer.Namespace)
		return err
	}

	return nil
}

// Deletes an Issuer i
func (r *CertManagerReconciler) DeleteIssuer(i certmanagerv1.Issuer) error {
	err := r.Client.Delete(context.TODO(), &i)
	if err != nil && !errors.IsNotFound(err) {
		logd.Error(err, "Failed to delete Issuer", "name:", res.Issuer.Name, "namespace:", res.Issuer.Namespace)
		return err
	}
	return nil
}

// get the deployment of cert-manager-controller
func (r *CertManagerReconciler) GetCertManagerControllerDeployment() error {
	deploy := &appsv1.Deployment{}
	deployName := "cert-manager-controller"
	deployNs := res.DeployNamespace

	err := r.Reader.Get(context.TODO(), types.NamespacedName{Name: deployName, Namespace: deployNs}, deploy)
	if err != nil {
		return err
	}

	// check the label of this deployment
	ControllorName := deploy.GetLabels()["app"]
	if ControllorName != "ibm-cert-manager-controller" {
		return fmt.Errorf("this controller don't have correct label")
	}

	return nil
}

// check this object has this label or not
func (r *CertManagerReconciler) CheckLabel(unstruct unstructured.Unstructured, labels map[string]string) bool {
	for k, v := range labels {
		if !r.HasLabel(unstruct, k) {
			return false
		}
		if unstruct.GetLabels()[k] != v {
			return false
		}
	}
	return true
}

func (r *CertManagerReconciler) HasLabel(cr unstructured.Unstructured, labelName string) bool {
	if cr.GetLabels() == nil {
		return false
	}
	if _, ok := cr.GetLabels()[labelName]; !ok {
		return false
	}
	return true
}

// GetObject get k8s resource with the unstructured object
func (r *CertManagerReconciler) GetObject(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	found := &unstructured.Unstructured{}
	found.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	err := r.Reader.Get(context.TODO(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, found)

	return found, err
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

// 1.create CertManager V1 Crds
// 2.add IBM label to this CRD
// 3.check existed crd is managed by ibm or not
// 4. if it is managed by ibm we can upgrade it
func (r *CertManagerReconciler) CreateOrUpdateV1CRDs() error {
	klog.Infof("Creating CertManager CRDs")
	labels := map[string]string{
		"app.kubernetes.io/instance":   "ibm-cert-manager-operator",
		"app.kubernetes.io/managed-by": "ibm-cert-manager-operator",
		"app.kubernetes.io/name":       "cert-manager",
	}
	var errMsg error
	CRDs := []string{
		res.CertificaterequestsCRD, res.CertificatesCRD, res.ClusterissuersCRD, res.IssuersCRD, res.OrdersCRD, res.ChallengesCRD,
	}
	for _, CRD := range CRDs {

		objects, err := YamlToObjects([]byte(CRD))
		if err != nil {
			return err
		}
		// obj is the object in yaml file
		// crd is the object in the cluster
		for _, obj := range objects {
			gvk := obj.GetObjectKind().GroupVersionKind()
			crd, err := r.GetObject(obj)
			version := obj.GetLabels()["app.kubernetes.io/version"]
			labels["app.kubernetes.io/version"] = version
			//this object not exist we need to create it
			if errors.IsNotFound(err) {
				klog.Infof("Creating resource with name: %s, namespace: %s, kind: %s, apiversion: %s/%s\n", obj.GetName(), obj.GetNamespace(), gvk.Kind, gvk.Group, gvk.Version)
				// add label locally
				if !r.CheckLabel(*obj, labels) {
					obj.SetLabels(labels)
				}
				if e := r.CreateObject(obj); e != nil {
					errMsg = e
				}
				continue
				// if the object exist
			} else if err == nil {
				//check if it haven't ibm label, skip it
				if !r.CheckLabel(*crd, managedbyLabel) && !r.CheckLabel(*crd, old_labels) {
					klog.Infof("this crd:%s is not managed by ibm-cert-manager, skip it", crd.GetName())
					continue
					//if it have ibm label
				} else {
					// update it
					r.UpdateResourse(obj, crd, labels)
				}
				// if can't getObject
			} else if err != nil {
				klog.Infof("can't get object:%s: %v", obj.GetName(), err)
				errMsg = err
				continue
			}
		}
	}
	return errMsg
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create certManager CRDs
	if err := r.CreateOrUpdateV1CRDs(); err != nil {
		klog.Errorf("Fail to create CRDs: %v", err)
		return err
	}

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
