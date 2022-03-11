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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/apis/cert-manager/v1"
	certmanagerv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/certmanager/v1alpha1"
	res "github.com/ibm/ibm-cert-manager-operator/controllers/resources"
)

// V1Alpha1AddLabelReconciler reconciles a Certificate object
type V1Alpha1AddLabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Certificate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *V1Alpha1AddLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logd = log.FromContext(ctx)

	reqLogger := logd.WithValues("req.Namespace", req.Namespace, "req.Name", req.Name)
	reqLogger.Info("Reconciling CertificateRefresh")

	v1cert := &certmanagerv1.Certificate{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, v1cert)
	if err == nil {
		// the corresponding v1 cert exists
		return ctrl.Result{}, nil
	} else if !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	cert := &certmanagerv1alpha1.Certificate{}
	err = r.Client.Get(context.TODO(), req.NamespacedName, cert)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get secret corresponding to the certificate
	secretInstance, err := r.getSecret(cert)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	oldLabelsMap := secretInstance.GetLabels()
	if oldLabelsMap == nil {
		oldLabelsMap = make(map[string]string)
	}

	if _, ok := oldLabelsMap[res.SecretWatchLabel]; !ok {
		oldLabelsMap[res.SecretWatchLabel] = ""
		secretInstance.SetLabels(oldLabelsMap)
	}

	r.updateSecret(secretInstance)
	return ctrl.Result{}, nil
}

// getSecret finds corresponding secret of the certmanagerv1alpha1 certificate
func (r *V1Alpha1AddLabelReconciler) getSecret(cert *certmanagerv1alpha1.Certificate) (*corev1.Secret, error) {
	secretName := cert.Spec.SecretName
	secret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: cert.Namespace}, secret)

	return secret, err
}

// updateSecret updates corresponding secret
func (r *V1Alpha1AddLabelReconciler) updateSecret(secret *corev1.Secret) error {
	return r.Client.Update(context.TODO(), secret)
}

// SetupWithManager sets up the controller with the Manager.
func (r *V1Alpha1AddLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create a new controller
	c, err := controller.New("addlabel-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to certmanagerv1alpha1 Certificates in the cluster
	err = c.Watch(&source.Kind{Type: &certmanagerv1alpha1.Certificate{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}
