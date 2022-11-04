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
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	secret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
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

	// found ca cert or ca secret
	foundCA := false
	// check this secret has refresh label or not
	// if this secret has refresh label
	if secret.GetLabels()[res.RefreshCALabel] == "true" {
		foundCA = true
	} else {
		// Get the certificate by this secret in the same namespace
		cert, err := r.getCertificateBySecret(secret)
		foundCert := true
		if err != nil {
			if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			logd.Info("Failed to find backing Certificate object for secret", "name:", secret.Name, "namespace:", secret.Namespace)
			foundCert = false
		}
		// Adding extra logic to check the duration of cs-ca-certificate. If no fields, then add the fields with default values
		// If fields exist, don't do anything
		if err := r.setCSCACertificateDuration(cert); err != nil {
			return ctrl.Result{}, err
		}

		// if we found this certificate in the same namespace
		if foundCert {
			// check if certificate is in list of CAs to refresh
			for _, caCert := range listOfCAs {
				if caCert.CertName == cert.Name && caCert.Namespace == cert.Namespace {
					foundCA = true
					break
				}
			}
			// check this certificate has refresh label or not
			if cert.Labels[res.RefreshCALabel] == "true" {
				foundCA = true
			}
		}
	}

	if !foundCA {
		//if certificate not in the list, disregard i.e. return and don't requeue
		logd.Info("Certificate Secret doesn't need its leaf certs refreshed. Disregarding.", "Secret.Name", secret.Name, "Secret.Namespace", secret.Namespace)
		return ctrl.Result{}, nil
	}

	logd.Info("Certificate Secret is a CA, its leaf should be refreshed", "Secret.Name", secret.Name, "Secret.Namespace", secret.Namespace)

	//Get tls.crt of the CA
	tlsValueOfCA := secret.Data["tls.crt"]

	// Fetch issuers
	issuers, err := r.findIssuersBasedOnCA(secret)
	if err != nil {
		return ctrl.Result{}, err
	}

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

	logd.Info("All leaf certificates refreshed for", "Secret.Name", secret.Name, "Secret.Namespace", secret.Namespace)
	return ctrl.Result{}, nil
}

// Get the certificate by secret in the same namespace
func (r *CertificateRefreshReconciler) getCertificateBySecret(secret *corev1.Secret) (*certmanagerv1.Certificate, error) {
	certName := secret.GetAnnotations()["cert-manager.io/certificate-name"]
	namespace := secret.GetNamespace()
	cert := &certmanagerv1.Certificate{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: certName, Namespace: namespace}, cert)

	return cert, err
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

func (r *CertificateRefreshReconciler) waitResourceReady(apiGroupVersion, kind string) error {
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
func (r *CertificateRefreshReconciler) ResourceExists(dc discovery.DiscoveryInterface, apiGroupVersion, kind string) (bool, error) {
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
func (r *CertificateRefreshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// wait for crd ready
	if err := r.waitResourceReady("cert-manager.io/v1", "Certificate"); err != nil {
		return err
	}

	// Create a new controller
	c, err := controller.New("certificaterefresh-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Certificates in the cluster
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}
