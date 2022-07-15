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

package certmanager

import (
	"context"
	"time"

	"github.com/ibm/ibm-cert-manager-operator/controllers/operator"
	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/v1apis/cert-manager/v1"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metaerrors "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	certmanagerv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/certmanager/v1alpha1"
)

// IssuerReconciler reconciles a Issuer object
type IssuerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=certmanager.k8s.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=certmanager.k8s.io,resources=issuers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=certmanager.k8s.io,resources=issuers/finalizers,verbs=update
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *IssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logd := log.FromContext(ctx)

	reqLogger := logd.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling Issuer")

	// Fetch the Issuer instance
	instance := &certmanagerv1alpha1.Issuer{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	//Check RHACM
	rhacmVersion, _, rhacmErr := operator.CheckRhacm(r.Client)
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

		if deployOperand < 0 {
			logd.Info("RHACM version is less than 2.3, so not reconciling Issuer")
			return ctrl.Result{}, nil
		}
	}

	reqLogger.Info("### DEBUG ### v1alpha1 Issuer created", "Issuer.Namespace", instance.Namespace, "Issuer.Name", instance.Name)

	reqLogger.Info("### DEBUG ### Creating v1 Issuer", "Issuer.Namespace", instance.Namespace, "Issuer.Name", instance.Name)

	annotations := instance.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["ibm-cert-manager-operator-generated"] = "true"

	v1Issuer := &certmanagerv1.Issuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Issuer",
			APIVersion: "cert-manager.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: annotations,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				ACME:       convertACME(instance.Spec.ACME),
				CA:         convertCA(instance.Spec.CA),
				Vault:      convertVault(instance.Spec.Vault),
				SelfSigned: convertSelfSigned(instance.Spec.SelfSigned),
				Venafi:     convertVenafi(instance.Spec.Venafi),
			},
		},
	}

	// Set the issuer v1alpha1 as the controller of the issuer v1
	if err := controllerutil.SetControllerReference(instance, v1Issuer, r.Scheme); err != nil {
		reqLogger.Error(err, "### DEBUG ### failed to set owner reference for %s", v1Issuer)
		return ctrl.Result{}, err
	}

	if err := r.Client.Create(context.TODO(), v1Issuer); err != nil {
		if errors.IsAlreadyExists(err) {
			existingIssuer := &certmanagerv1.Issuer{}
			if err := r.Client.Get(context.TODO(), types.NamespacedName{Namespace: v1Issuer.Namespace, Name: v1Issuer.Name}, existingIssuer); err != nil {
				reqLogger.Error(err, "### DEBUG ### Failed to get v1 Issuer")
				return ctrl.Result{}, err
			}
			if !equality.Semantic.DeepEqual(v1Issuer.Labels, existingIssuer.Labels) || !equality.Semantic.DeepEqual(v1Issuer.Spec, existingIssuer.Spec) {
				v1Issuer.SetResourceVersion(existingIssuer.GetResourceVersion())
				v1Issuer.SetAnnotations(existingIssuer.GetAnnotations())
				if err := r.Client.Update(context.TODO(), v1Issuer); err != nil {
					reqLogger.Error(err, "### DEBUG ### Failed to update v1 Issuer")
					return ctrl.Result{}, err
				}
				reqLogger.Info("### DEBUG #### Updated v1 Issuer")
			}

			reqLogger.Info("### DEBUG ### Converting Issuer status")
			status := convertIssuerStatus(existingIssuer.Status)
			instance.Status = status
			reqLogger.Info("### DEBUG ### Updating v1alpha1 Issuer status")
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "### DEBUG ### error updating")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
		reqLogger.Error(err, "### DEBUG ### Failed to create v1 Issuer")
		return ctrl.Result{}, err
	}

	reqLogger.Info("### DEBUG #### Created v1 Issuer")

	return ctrl.Result{}, nil
}

type ignoreStatusPredicate struct{}

func (i ignoreStatusPredicate) Create(e event.CreateEvent) bool {
	return true
}

func (i ignoreStatusPredicate) Delete(e event.DeleteEvent) bool {
	return false
}

func (i ignoreStatusPredicate) Update(e event.UpdateEvent) bool {
	return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
}

func (i ignoreStatusPredicate) Generic(e event.GenericEvent) bool {
	return false
}

func (r *IssuerReconciler) waitResourceReady(apiGroupVersion, kind string) error {
	klog.Infof("wait for resource ready")
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Errorf("Failed to get config: %v", err)
		return err
	}
	dc := discovery.NewDiscoveryClientForConfigOrDie(cfg)
	if err := utilwait.PollImmediate(time.Second*10, time.Minute*5, func() (done bool, err error) {
		exist, err := r.ResourceExists(dc, apiGroupVersion, kind)
		if err != nil {
			return exist, err
		}
		if !exist {
			klog.Infof("waiting for resource ready with kind: %s, apiGroupVersion: %s", kind, apiGroupVersion)
		}
		return exist, nil
	}); err != nil {
		return err
	}
	return nil
}

// ResourceExists returns true if the given resource kind exists
// in the given api groupversion
func (r *IssuerReconciler) ResourceExists(dc discovery.DiscoveryInterface, apiGroupVersion, kind string) (bool, error) {
	_, apiLists, err := dc.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == apiGroupVersion {
			for _, r := range apiList.APIResources {
				if r.Kind == kind {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// wait for crd ready
	if err := r.waitResourceReady("cert-manager.io/v1", "Issuer"); err != nil {
		return err
	}
	// Create a new controller
	c, err := controller.New("issuer-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Issuer
	err = c.Watch(&source.Kind{Type: &certmanagerv1alpha1.Issuer{}}, &handler.EnqueueRequestForObject{}, ignoreStatusPredicate{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner Issuer
	err = c.Watch(&source.Kind{Type: &certmanagerv1.Issuer{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &certmanagerv1alpha1.Issuer{},
	})
	if err != nil {
		return err
	}

	return nil
}
