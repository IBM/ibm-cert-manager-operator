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
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	certmanagerv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/certmanager/v1alpha1"
	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-cert-manager-operator/controllers/resources"
	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/v1apis/cert-manager/v1"
)

var logd = log.Log.WithName("controller_certificaterefresh")

// CertificateReconciler reconciles a Certificate object
type CertificateRefreshReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Certificate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CertificateRefreshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logd = log.FromContext(ctx)

	reqLogger := logd.WithValues("req.Namespace", req.Namespace, "req.Name", req.Name)
	reqLogger.Info("Reconciling CertificateRefresh")

	// Get the certificate that invoked reconciliation is a CA in the listOfCAs

	cert := &certmanagerv1.Certificate{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, cert)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// trigger conversion controller to handle mixing v1 CA Certificates and
	// v1alpha1 leaf Certificates
	// if cert.Labels[res.OperatorGeneratedAnno] == "" && cert.Labels[res.ProperV1Label] == "" {
	// 	v1alpha1 := &certmanagerv1alpha1.Certificate{}
	// 	if err := r.Client.Get(context.TODO(), req.NamespacedName, v1alpha1); err != nil {
	// 		if errors.IsNotFound(err) {
	// 			// Request object not found, could have been deleted after reconcile req
	// 			// Return and don't requeue
	// 			reqLogger.Info("Could not find v1alpha1 certificate")
	// 			return ctrl.Result{}, nil
	// 		}
	// 		return ctrl.Result{}, err
	// 	}
	// 	reqLogger.Info("Emptying v1alpha1 Cert status")
	// 	v1alpha1.Status = certmanagerv1alpha1.CertificateStatus{}
	// 	if err := r.Client.Update(context.TODO(), v1alpha1); err != nil {
	// 		reqLogger.Error(err, "failed to empty v1alpha1 status")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	// Adding extra logic to check the duration of cs-ca-certificate. If no fields, then add the fields with default values
	// If fields exist, don't do anything
	if err := r.setCSCACertificateDuration(cert); err != nil {
		return ctrl.Result{}, err
	}

	// Fetch the CertManager instance to check the enableCertRefresh flag
	instance := &operatorv1alpha1.CertManager{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: res.CertManagerInstanceName, Namespace: ""}, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			logd.Info("CR instance not found, don't requeue")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logd.Error(err, "Error reading the object - requeue the request")
		return ctrl.Result{}, err
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
		logd.Info("Flag EnableCertRefresh is set to false, don't requeue")
		return ctrl.Result{}, nil
	}

	//set the list of CAs that need their leaf certs refreshed
	listOfCAs = r.buildDefaultCAList()
	listOfCAs = append(listOfCAs, instance.Spec.RefreshCertsBasedOnCA...)

	logd.V(2).Info("refreshCertsBasedOnCA list: ", "", listOfCAs)

	if len(listOfCAs) == 0 {
		logd.Info("List of CAs empty. No leaf certificates to refresh")
		return ctrl.Result{}, nil
	}

	logd.Info("Flag EnableCertRefresh is set to true!")

	found := false
	for _, caCert := range listOfCAs {
		if caCert.CertName == cert.Name && caCert.Namespace == cert.Namespace {
			found = true
			break
		}
	}

	if cert.Labels[res.RefreshCALabel] == "true" {
		found = true
	}

	if !found {
		//if certificate not in the list, disregard i.e. return and don't requeue
		logd.Info("Certificate doesn't need its leaf certs refreshed. Disregarding.", "Certificate.Name", cert.Name, "Certificate.Namespace", cert.Namespace)
		return ctrl.Result{}, nil
	}

	logd.Info("Certificate is a CA, its leaf should be refreshed", "Certificate.Name", cert.Name, "Certificate.Namespace", cert.Namespace)

	// Get secret corresponding to the CA certificate
	caSecret, err := r.getSecret(cert)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	//Get tls.crt of the CA
	tlsValueOfCA := caSecret.Data["tls.crt"]

	// Fetch issuers
	issuers, err := r.findIssuersBasedOnCA(caSecret)
	if err != nil {
		return ctrl.Result{}, err
	}

	//nolint
	//TODO: Add clusterissuer to api
	// // Fetch clusterissuers
	// clusterissuers, err := r.findClusterIssuersBasedOnCA(caSecret)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// // Fetch all the secrets of leaf certificates issued by these issuers/clusterissuers
	var leafSecrets []*corev1.Secret

	v1LeafCerts, err := r.findV1Certs(issuers)
	if err != nil {
		logd.Error(err, "Error reading the leaf certificates for issuer - requeue the request")
		return ctrl.Result{}, err
	}

	leafSecrets, err = r.findLeafSecrets(v1LeafCerts)
	if err != nil {
		logd.Error(err, "Error finding secrets from v1 leaf certificates - requeue the request")
		return ctrl.Result{}, err
	}

	v1alpha1Leaves, err := r.findV1Alpha1Certs(issuers, v1LeafCerts...)
	if err != nil {
		return ctrl.Result{}, err
	}
	logd.V(2).Info("List of v1alpha1 leaves for refresh", "v1alpha1 certs", v1alpha1Leaves)

	//nolint
	//TODO: Add clusterissuer to api
	// allNamespaces, err := r.getAllNamespaces()
	// if err != nil {
	// 	logd.Error(err, "Error listing all namespaces - requeue the request")
	// 	return ctrl.Result{}, err
	// }

	//nolint
	//TODO: Add clusterissuer to api
	// for _, clusterissuer := range clusterissuers {
	// 	for _, ns := range allNamespaces.Items {
	// 		clusterLeafSecrets, err := r.findLeafSecrets(clusterissuer.Name, ns.Name)
	// 		if err != nil {
	// 			logd.Error(err, "Error reading the leaf certificates for clusterissuer - requeue the request")
	// 			return ctrl.Result{}, err
	// 		}
	// 		leafSecrets = append(leafSecrets, clusterLeafSecrets...)
	// 	}
	// }

	// Compare ca.crt in leaf with tls.crt of CA
	// If the values don't match, delete the secret; if error, requeue else don't requeue
	for _, leafSecret := range leafSecrets {
		if string(leafSecret.Data["ca.crt"]) != string(tlsValueOfCA) {
			logd.Info("Deleting leaf secret " + leafSecret.Name + " as ca.crt value has changed")
			if err := r.Client.Delete(context.TODO(), leafSecret); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return ctrl.Result{}, err
			}
		}
	}

	// clear status of v1alpha1 leaf certs
	logd.Info("Refreshing v1alpha1 leaf certs")
	for _, c := range v1alpha1Leaves {
		c.Status = certmanagerv1alpha1.CertificateStatus{}
		if err := r.Client.Update(context.TODO(), &c); err != nil {
			return ctrl.Result{}, err
		}
		secret := &corev1.Secret{}
		if err := r.Client.Get(context.TODO(), types.NamespacedName{
			Namespace: c.Namespace,
			Name:      c.Spec.SecretName,
		}, secret); err != nil {
			if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
		}
		if err := r.Client.Delete(context.TODO(), secret); err != nil {
			return ctrl.Result{}, err
		}
	}

	logd.Info("All leaf certificates refreshed for", "Certificate.Name", cert.Name, "Certificate.Namespace", cert.Namespace)
	return ctrl.Result{}, nil
}

//setCSCACertificateDuration sets duration of cs-ca-certificate to 2 years
func (r *CertificateRefreshReconciler) setCSCACertificateDuration(cert *certmanagerv1.Certificate) error {

	if cert.Name != res.CSCACertName || cert.Namespace != res.DeployNamespace {
		return nil
	}

	if cert.Spec.Duration != nil && cert.Spec.RenewBefore != nil {
		return nil
	}

	patch := client.MergeFrom(cert.DeepCopy())
	cert.Spec.Duration = &metav1.Duration{Duration: time.Hour * 24 * 365 * 2}
	cert.Spec.RenewBefore = &metav1.Duration{Duration: time.Hour * 24 * 30}

	if err := r.Client.Patch(context.TODO(), cert, patch); err != nil {
		return err
	}

	logd.Info("CS CA Certificate duration set to 2 years and renewal set to 30 days before expiration ")

	// Get secret corresponding to the CA certificate
	cscaSecret, err := r.getSecret(cert)
	if err == nil {
		//delete secret to refresh it with the new duration: bug in v0.10 cert-manager; resolved from v0.15
		err = r.Client.Delete(context.TODO(), cscaSecret)
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

		if err := r.Client.Patch(context.TODO(), cert, patch); err != nil {
			logd.Info("Error patching the certificate")
			return err
		}
		logd.Info("Error retrieving/deleting cs-ca-certificate-secret; resetting duration/renewBefore ")
		return err
	}

	return nil
}

// getSecret finds corresponding secret of the certificate
func (r *CertificateRefreshReconciler) getSecret(cert *certmanagerv1.Certificate) (*corev1.Secret, error) {
	secretName := cert.Spec.SecretName
	secret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: cert.Namespace}, secret)

	return secret, err
}

func (r *CertificateRefreshReconciler) buildDefaultCAList() []operatorv1alpha1.CACertificate {
	defaultCAs := make([]operatorv1alpha1.CACertificate, 0)

	logd.Info(fmt.Sprintf("Finding all namespaces where %s is deployed", res.ProductName))

	odlmDeployments := &appsv1.DeploymentList{}
	err := r.Client.List(context.TODO(), odlmDeployments, &client.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{
			"metadata.name": res.OdlmDeploymentName,
		}),
	})
	if err != nil {
		logd.Error(err, "Error listing ODLM deployments")
		return defaultCAs
	}

	logd.Info("Building default list of CA certificates for leaf certificate refresh")

	for _, d := range odlmDeployments.Items {
		for _, name := range res.DefaultCANames {
			defaultCAs = append(defaultCAs, operatorv1alpha1.CACertificate{
				CertName:  name,
				Namespace: d.GetNamespace(),
			})
		}
	}

	return defaultCAs
}

// findIssuersBasedOnCA finds issuers that are based on the given CA secret
func (r *CertificateRefreshReconciler) findIssuersBasedOnCA(caSecret *corev1.Secret) ([]certmanagerv1.Issuer, error) {

	var issuers []certmanagerv1.Issuer

	issuerList := &certmanagerv1.IssuerList{}
	err := r.Client.List(context.TODO(), issuerList, &client.ListOptions{Namespace: caSecret.Namespace})
	if err == nil {
		for _, issuer := range issuerList.Items {
			if issuer.Spec.CA != nil && issuer.Spec.CA.SecretName == caSecret.Name {
				issuers = append(issuers, issuer)
			}
		}
	}

	return issuers, err
}

func (r *CertificateRefreshReconciler) findV1Certs(issuers []certmanagerv1.Issuer) ([]certmanagerv1.Certificate, error) {
	var leafCerts []certmanagerv1.Certificate
	for _, i := range issuers {
		certList := &certmanagerv1.CertificateList{}
		err := r.Client.List(context.TODO(), certList, &client.ListOptions{Namespace: i.Namespace})
		if err != nil {
			return leafCerts, err
		}

		for _, c := range certList.Items {
			if c.Spec.IssuerRef.Name == i.Name {
				leafCerts = append(leafCerts, c)
			}
		}
	}
	return leafCerts, nil
}

// findLeafSecrets finds issuers that are based on the given CA secret
func (r *CertificateRefreshReconciler) findLeafSecrets(v1Certs []certmanagerv1.Certificate) ([]*corev1.Secret, error) {

	var leafSecrets []*corev1.Secret

	for _, cert := range v1Certs {
		leafSecret, err := r.getSecret(&cert)
		if err != nil {
			if errors.IsNotFound(err) {
				logd.V(2).Info("Secret not found for cert " + cert.Name)
				continue
			}
			return leafSecrets, err
		}
		leafSecrets = append(leafSecrets, leafSecret)
	}

	return leafSecrets, nil
}

// findV1Alpha1Certs Finds all the v1alpha1 Certificates which have not been
// converted to avoid deleting the same certificate secret twice
func (r *CertificateRefreshReconciler) findV1Alpha1Certs(issuers []certmanagerv1.Issuer, v1Certs ...certmanagerv1.Certificate) ([]certmanagerv1alpha1.Certificate, error) {
	certs := &certmanagerv1alpha1.CertificateList{}
	var v1alpha1Certs []certmanagerv1alpha1.Certificate

	issuerNames := []string{}
	for _, i := range issuers {
		issuerNames = append(issuerNames, i.Name)
	}
	requirement, err := labels.NewRequirement("certmanager.k8s.io/issuer-name", selection.In, issuerNames)
	if err != nil {
		return v1alpha1Certs, err
	}
	selector := labels.NewSelector().Add(*requirement)

	if err := r.Client.List(context.TODO(), certs, &client.ListOptions{
		LabelSelector: selector,
	}); err != nil {
		return v1alpha1Certs, err
	}

	for _, c := range certs.Items {
		found := false
		for _, v := range v1Certs {
			if c.Name == v.Name && c.Namespace == v.Namespace {
				found = true
				break
			}
		}
		if !found {
			v1alpha1Certs = append(v1alpha1Certs, c)
		}
	}

	return v1alpha1Certs, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateRefreshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create a new controller
	c, err := controller.New("certificaterefresh-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Certificates in the cluster
	err = c.Watch(&source.Kind{Type: &certmanagerv1.Certificate{}}, &handler.EnqueueRequestForObject{}, isCACertificatePredicate{})
	if err != nil {
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
	reqCert := (e.ObjectOld).(*certmanagerv1.Certificate)
	if !reqCert.Spec.IsCA {
		return false
	}

	return e.ObjectOld != e.ObjectNew
}

// Create implements default CreateEvent filter for validating if object is a CA
// Intended to be used with certificates.

func (isCACertificatePredicate) Create(e event.CreateEvent) bool {
	reqCert := (e.Object).(*certmanagerv1.Certificate)

	return reqCert.Spec.IsCA
}

func (isCACertificatePredicate) Delete(e event.DeleteEvent) bool {
	return false
}

func (isCACertificatePredicate) Generic(e event.GenericEvent) bool {
	return false
}
