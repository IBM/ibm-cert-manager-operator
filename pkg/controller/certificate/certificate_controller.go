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

package certificate

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1"
	certmanagerv1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1alpha1"
)

var log = logf.Log.WithName("controller_certificate")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Certificate Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCertificate{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("certificate-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Certificate
	err = c.Watch(&source.Kind{Type: &certmanagerv1alpha1.Certificate{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Certificate
	err = c.Watch(&source.Kind{Type: &certmanagerv1.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &certmanagerv1alpha1.Certificate{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCertificate implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCertificate{}

// ReconcileCertificate reconciles a Certificate object
type ReconcileCertificate struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Certificate object and makes changes based on the state read
// and what is in the Certificate.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCertificate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Certificate")

	// Fetch the Certificate instance
	instance := &certmanagerv1alpha1.Certificate{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reqLogger.Info("### DEBUG ### v1alpha1 Certificate created", "Certificate.Namespace", instance.Namespace, "Certificate.Name", instance.Name)

	reqLogger.Info("### DEBUG ### Creating v1 Certificate", "Certificate.Namespace", instance.Namespace, "Certificate.Name", instance.Name)

	annotations := instance.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["ibm-cert-manager-operator-generated"] = "true"

	certificate := &certmanagerv1.Certificate{
		TypeMeta: metav1.TypeMeta{Kind: "Certificate", APIVersion: "cert-manager.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      instance.Labels,
			Annotations: annotations,
		},
		Spec: certmanagerv1.CertificateSpec{
			Subject:               convertSubject(instance.Spec.Organization),
			CommonName:            instance.Spec.CommonName,
			Duration:              instance.Spec.Duration,
			RenewBefore:           instance.Spec.RenewBefore,
			DNSNames:              instance.Spec.DNSNames,
			IPAddresses:           instance.Spec.IPAddresses,
			URIs:                  nil,
			EmailAddresses:        nil,
			SecretName:            instance.Spec.SecretName,
			Keystores:             nil,
			IssuerRef:             convertIssuerRef(instance.Spec.IssuerRef),
			IsCA:                  instance.Spec.IsCA,
			Usages:                convertUsages(instance.Spec.Usages),
			PrivateKey:            convertPrivateKey(instance.Spec),
			EncodeUsagesInRequest: nil,
			RevisionHistoryLimit:  nil,
		},
	}
	// Set the certificate v1alpha1 as the controller of the certificate v1
	if err := controllerutil.SetControllerReference(instance, certificate, r.scheme); err != nil {
		reqLogger.Error(err, "### DEBUG ### failed to set Owner reference for %s", certificate)
		return reconcile.Result{}, err
	}

	if err := r.client.Create(context.TODO(), certificate); err != nil {
		if errors.IsAlreadyExists(err) {
			existingCertificate := &certmanagerv1.Certificate{}
			if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: certificate.Namespace, Name: certificate.Name}, existingCertificate); err != nil {
				reqLogger.Error(err, "### DEBUG ### Failed to get v1 Certificate")
				return reconcile.Result{}, err
			}
			if !equality.Semantic.DeepEqual(certificate.Labels, existingCertificate.Labels) || !equality.Semantic.DeepEqual(certificate.Spec, existingCertificate.Spec) {
				certificate.SetResourceVersion(existingCertificate.GetResourceVersion())
				certificate.SetAnnotations(existingCertificate.GetAnnotations())
				if err := r.client.Update(context.TODO(), certificate); err != nil {
					reqLogger.Error(err, "### DEBUG ### Failed to update v1 Certificate")
					return reconcile.Result{}, err
				}
				reqLogger.Info("### DEBUG #### Updated v1 Certificate")
			}

			reqLogger.Info("### DEBUG ### Converting status")
			status := convertStatus(existingCertificate.Status)
			instance.Status = status
			reqLogger.Info("### DEBUG ### Updating v1alpha1 status")
			if err := r.client.Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "### DEBUG ### error patching")
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "### DEBUG ### Failed to create v1 Certificate")
		return reconcile.Result{}, err
	}

	reqLogger.Info("### DEBUG #### Created v1 Certificate")

	return reconcile.Result{}, nil
}
