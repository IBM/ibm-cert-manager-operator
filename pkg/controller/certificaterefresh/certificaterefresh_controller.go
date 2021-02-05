//
// Copyright 2021 IBM Corporation
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

package certificaterefresh

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-cert-manager-operator/pkg/resources"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_certificaterefresh")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new CertificateRefresh Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCertificateRefresh{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("certificaterefresh-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Certificates in the cluster
	err = c.Watch(&source.Kind{Type: &certmgr.Certificate{}}, &handler.EnqueueRequestForObject{}, isCACertificatePredicate{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCertificateRefresh implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCertificateRefresh{}

// ReconcileCertificateRefresh reconciles a CertificateRefresh object
type ReconcileCertificateRefresh struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Certificate object and makes changes based on the state read
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCertificateRefresh) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CertificateRefresh")

	// Get the certificate that invoked reconciliation is a CA in the listOfCAs

	cert := &certmgr.Certificate{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cert)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Adding extra logic to check the duration of cs-ca-certificate. If no fields, then add the fields with default values
	// If fields exist, don't do anything
	if cert.Name == res.CSCACertName && cert.Namespace == res.DeployNamespace && (cert.Spec.Duration == nil || cert.Spec.RenewBefore == nil) {
		if err := r.setCSCACertificateDuration(cert); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Fetch the CertManager instance to check the enableCertRefresh flag
	instance := &operatorv1alpha1.CertManager{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.CertManagerInstanceName, Namespace: ""}, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("CR instance not found, don't requeue")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Error reading the object - requeue the request")
		return reconcile.Result{}, err
	}

	enableCertRefresh := false
	var listOfCAs []operatorv1alpha1.CACertificate

	if instance.Spec.EnableCertRefresh == nil {
		//default value
		enableCertRefresh = res.DefaultEnableCertRefresh
	} else {
		enableCertRefresh = *instance.Spec.EnableCertRefresh
	}

	if !enableCertRefresh {
		log.Info("Flag EnableCertRefresh is set to false, don't requeue")
		return reconcile.Result{}, nil
	}

	//set the list of CAs that need their leaf certs refreshed
	listOfCAs = res.DefaultCAList
	listOfCAs = append(listOfCAs, instance.Spec.RefreshCertsBasedOnCA...)

	if len(listOfCAs) == 0 {
		log.Info("List of CAs empty. No leaf certificates to refresh")
		return reconcile.Result{}, nil
	}

	log.Info("Flag EnableCertRefresh is set to true!")

	found := false
	for _, caCert := range listOfCAs {
		if caCert.CertName == cert.Name && caCert.Namespace == cert.Namespace {
			found = true
			break
		}
	}

	if !found {
		//if certificate not in the list, disregard i.e. return and don't requeue
		log.Info("Certificate is not a CA/doesn't need its leaf certs refreshed. Disregarding.", "Certificate.Name", cert.Name, "Certificate.Namespace", cert.Namespace)
		return reconcile.Result{}, nil
	}

	log.Info("Certificate is a CA, its leaf should be refreshed", "Certificate.Name", cert.Name, "Certificate.Namespace", cert.Namespace)

	// Get secret corresponding to the CA certificate
	caSecret, err := r.getSecret(cert)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	//Get tls.crt of the CA
	tlsValueOfCA := caSecret.Data["tls.crt"]

	// Fetch issuers
	issuers, err := r.findIssuersBasedOnCA(caSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Fetch clusterissuers
	clusterissuers, err := r.findClusterIssuersBasedOnCA(caSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	// // Fetch all the secrets of leaf certificates issued by these issuers/clusterissuers
	var leafSecrets []*corev1.Secret

	for _, issuer := range issuers {
		leafSecrets, err = r.findLeafSecrets(issuer.Name, issuer.Namespace)
		if err != nil {
			log.Error(err, "Error reading the leaf certificates for issuer - requeue the request")
			return reconcile.Result{}, err
		}
	}

	allNamespaces, err := r.getAllNamespaces()
	if err != nil {
		log.Error(err, "Error listing all namespaces - requeue the request")
		return reconcile.Result{}, err
	}

	for _, clusterissuer := range clusterissuers {
		for _, ns := range allNamespaces.Items {
			clusterLeafSecrets, err := r.findLeafSecrets(clusterissuer.Name, ns.Name)
			if err != nil {
				log.Error(err, "Error reading the leaf certificates for clusterissuer - requeue the request")
				return reconcile.Result{}, err
			}
			leafSecrets = append(leafSecrets, clusterLeafSecrets...)
		}
	}

	// Compare ca.crt in leaf with tls.crt of CA
	// If the values don't match, delete the secret; if error, requeue else don't requeue
	for _, leafSecret := range leafSecrets {
		if string(leafSecret.Data["ca.crt"]) != string(tlsValueOfCA) {
			log.Info("Deleting leaf secret " + leafSecret.Name + " as ca.crt value has changed")
			if err := r.client.Delete(context.TODO(), leafSecret); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return reconcile.Result{}, err
			}
		}
	}

	log.Info("All leaf certificates refreshed for", "Certificate.Name", cert.Name, "Certificate.Namespace", cert.Namespace)
	return reconcile.Result{}, nil
}

// getSecret finds corresponding secret of the certificate
func (r *ReconcileCertificateRefresh) getSecret(cert *certmgr.Certificate) (*corev1.Secret, error) {
	secretName := cert.Spec.SecretName
	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: cert.Namespace}, secret)

	return secret, err
}

// findIssuersBasedOnCA finds issuers that are based on the given CA secret
func (r *ReconcileCertificateRefresh) findIssuersBasedOnCA(caSecret *corev1.Secret) ([]certmgr.Issuer, error) {

	var issuers []certmgr.Issuer

	issuerList := &certmgr.IssuerList{}
	err := r.client.List(context.TODO(), issuerList, &client.ListOptions{Namespace: caSecret.Namespace})
	if err == nil {
		for _, issuer := range issuerList.Items {
			if issuer.Spec.CA != nil && issuer.Spec.CA.SecretName == caSecret.Name {
				issuers = append(issuers, issuer)
			}
		}
	}

	return issuers, err
}

// findClusterIssuersBasedOnCA finds issuers that are based on the given CA secret
func (r *ReconcileCertificateRefresh) findClusterIssuersBasedOnCA(caSecret *corev1.Secret) ([]certmgr.ClusterIssuer, error) {

	var clusterissuers []certmgr.ClusterIssuer

	clusterIssuerList := &certmgr.ClusterIssuerList{}
	err := r.client.List(context.TODO(), clusterIssuerList, &client.ListOptions{})

	if err == nil {
		for _, cissuer := range clusterIssuerList.Items {
			if cissuer.Spec.CA != nil && cissuer.Spec.CA.SecretName == caSecret.Name {
				clusterissuers = append(clusterissuers, cissuer)
			}
		}
	}

	return clusterissuers, err
}

// findLeafSecrets finds issuers that are based on the given CA secret
func (r *ReconcileCertificateRefresh) findLeafSecrets(issuedBy string, namespace string) ([]*corev1.Secret, error) {

	var leafSecrets []*corev1.Secret

	certList := &certmgr.CertificateList{}
	err := r.client.List(context.TODO(), certList, &client.ListOptions{Namespace: namespace})

	if err == nil {
		for _, cert := range certList.Items {
			if cert.Spec.IssuerRef.Name == issuedBy {
				leafSecret, err := r.getSecret(&cert)
				if err != nil {
					if errors.IsNotFound(err) {
						log.V(2).Info("Secret not found for cert " + cert.Name)
						continue
					}
					break
				}
				leafSecrets = append(leafSecrets, leafSecret)
			}
		}
	}

	return leafSecrets, err
}

// getAllNamespaces finds all namespaces in the cluster
func (r *ReconcileCertificateRefresh) getAllNamespaces() (*corev1.NamespaceList, error) {

	nsList := &corev1.NamespaceList{}
	err := r.client.List(context.TODO(), nsList, &client.ListOptions{})

	return nsList, err
}

//setCSCACertificateDuration sets duration of cs-ca-certificate to 2 years
func (r *ReconcileCertificateRefresh) setCSCACertificateDuration(cert *certmgr.Certificate) error {

	patch := client.MergeFrom(cert.DeepCopy())
	cert.Spec.Duration = &metav1.Duration{Duration: time.Hour * 24 * 365 * 2}
	cert.Spec.RenewBefore = &metav1.Duration{Duration: time.Hour * 24 * 30}

	if err := r.client.Patch(context.TODO(), cert, patch); err != nil {
		return err
	}

	log.Info("CS CA Certificate duration set to 2 years and renewal set to 30 days before expiration ")

	// Get secret corresponding to the CA certificate
	cscaSecret, err := r.getSecret(cert)
	if err == nil {
		//delete secret to refresh it with the new duration: bug in v0.10 cert-manager; resolved from v0.15
		err = r.client.Delete(context.TODO(), cscaSecret)
	}

	if err != nil {
		if errors.IsNotFound(err) {
			// secret is not created yet, or secret is deleted; will be created later and it will pick the new duration/renewBefore values set
			//no need to requeue
			return nil
		}
		// updating it to nil and requeueing so that we again attempt to set the duration/renewBefore and delete the secret
		patch := client.MergeFrom(cert.DeepCopy())
		cert.Spec.Duration = nil
		cert.Spec.RenewBefore = nil

		if err := r.client.Patch(context.TODO(), cert, patch); err != nil {
			log.Info("Error patching the certificate")
			return err
		}
		log.Info("Error retrieving/deleting cs-ca-certificate-secret; resetting duration/renewBefore ")
		return err
	}

	return nil
}

// isCACertificatePredicate implements a predicate verifying that
// a certificate is a CA certificate. This only applies to Create and Update events. Deletes
// and Generics should not make it to the work queue.
type isCACertificatePredicate struct{}

// Update implements default UpdateEvent filter for validating if object has the
// `isCA: true` which helps identify that it is a CA certificate.
// Intended to be used with certificates.
func (isCACertificatePredicate) Update(e event.UpdateEvent) bool {
	reqCert := (e.ObjectOld).(*certmgr.Certificate)
	if !reqCert.Spec.IsCA {
		return false
	}

	return e.ObjectOld != e.ObjectNew
}

// Create implements default CreateEvent filter for validating if object is a CA
// Intended to be used with certificates.

func (isCACertificatePredicate) Create(e event.CreateEvent) bool {
	reqCert := (e.Object).(*certmgr.Certificate)

	return reqCert.Spec.IsCA
}

func (isCACertificatePredicate) Delete(e event.DeleteEvent) bool {
	return false
}

func (isCACertificatePredicate) Generic(e event.GenericEvent) bool {
	return false
}
