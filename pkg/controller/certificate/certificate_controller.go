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
	"time"

	"golang.org/x/mod/semver"

	admRegv1 "k8s.io/api/admissionregistration/v1"

	corev1 "k8s.io/api/core/v1"
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
	"github.com/ibm/ibm-cert-manager-operator/pkg/controller/certmanager"
	"github.com/ibm/ibm-cert-manager-operator/pkg/resources"
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

const t = "true"

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

	//Check RHACM
	rhacmVersion, _, rhacmErr := certmanager.CheckRhacm(r.client)
	if rhacmErr != nil {
		// will continue since RHACM most likely not installed
		// logging error in case there is some other issue
		log.Error(rhacmErr, "Error trying to check for RHACM")
	}
	if rhacmVersion != "" {
		rhacmVersion = "v" + rhacmVersion
		deployOperand := semver.Compare(rhacmVersion, "v2.3")

		if deployOperand < 0 {
			log.Info("RHACM version is less than 2.3, so not reconciling Certificate")
			return reconcile.Result{}, nil
		}
	}

	reqLogger.Info("purging old v1 Certs")
	if err := r.purgeOldV1(); err != nil {
		reqLogger.Error(err, "failed to remove all old v1 Certificates ")
		return reconcile.Result{}, err
	}

	reqLogger.V(2).Info("Initializing v1 Certificate from v1alpha1", "Certificate.Namespace", instance.Namespace, "Certificate.Name", instance.Name)

	annotations := instance.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[resources.OperatorGeneratedAnno] = t

	labels := instance.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[resources.ProperV1Label] = t

	dnsNames, ipAddresses := sanitizeDNSNames(instance.Spec.DNSNames)

	certificate := &certmanagerv1.Certificate{
		TypeMeta: metav1.TypeMeta{Kind: "Certificate", APIVersion: "cert-manager.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Name,
			Namespace:   instance.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: certmanagerv1.CertificateSpec{
			Subject:               convertSubject(instance.Spec.Organization),
			CommonName:            convertCommonName(instance.Spec.CommonName, instance.Spec.DNSNames),
			Duration:              instance.Spec.Duration,
			RenewBefore:           instance.Spec.RenewBefore,
			DNSNames:              dnsNames,
			IPAddresses:           convertIPAddresses(instance.Spec.IPAddresses, ipAddresses),
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
		reqLogger.Error(err, "failed to set Owner reference for %s", certificate)
		return reconcile.Result{}, err
	}

	reqLogger.Info("Getting certificate secret")
	secret := &corev1.Secret{}
	nsname := types.NamespacedName{Name: instance.Spec.SecretName, Namespace: instance.Namespace}
	err = r.client.Get(context.TODO(), nsname, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No secret found, continuing")
			secret = nil
		} else {
			return reconcile.Result{}, err
		}
	}

	if isExpired(instance, secret) {
		reqLogger.Info("v1alpha1 Certificate is expired, creating v1 version")
		if err := r.client.Create(context.TODO(), certificate); err != nil {
			if errors.IsAlreadyExists(err) {
				existingCertificate := &certmanagerv1.Certificate{}
				if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: certificate.Namespace, Name: certificate.Name}, existingCertificate); err != nil {
					reqLogger.Error(err, "failed to get v1 Certificate")
					return reconcile.Result{}, err
				}
				if !equality.Semantic.DeepEqual(certificate.Labels, existingCertificate.Labels) || !equality.Semantic.DeepEqual(certificate.Spec, existingCertificate.Spec) {
					certificate.SetResourceVersion(existingCertificate.GetResourceVersion())
					certificate.SetAnnotations(existingCertificate.GetAnnotations())
					if err := r.client.Update(context.TODO(), certificate); err != nil {
						reqLogger.Error(err, "failed to update v1 Certificate")
						return reconcile.Result{}, err
					}
					reqLogger.Info("Updated v1 Certificate")
				}

				reqLogger.Info("Converting Certificate status")
				status := convertStatus(existingCertificate.Status)
				instance.Status = status
				reqLogger.Info("Updating v1alpha1 Certificate status")
				if err := r.client.Update(context.TODO(), instance); err != nil {
					reqLogger.Error(err, "error updating status")
					return reconcile.Result{}, err
				}
			} else {
				reqLogger.Error(err, "failed to create v1 Certificate")
				return reconcile.Result{}, err
			}
		}

		if err := r.updateWebhooks(instance.Namespace + "/" + instance.Name); err != nil {
			return reconcile.Result{}, err
		}

		// leaf certificate refresh logic enabled by default in conversion logic
		// otherwise services can be broken when v1alpha1 Certificates expire
		// and are automatically converted to v1
		if instance.Spec.IsCA {
			reqLogger.Info("CA Certificate has refreshed from upgrade, converting v1alpha1 leaf Certificates")
			caSecret, err := r.getSecret(instance)
			if err != nil {
				if errors.IsNotFound(err) {
					return reconcile.Result{}, nil
				}
				return reconcile.Result{}, err
			}

			issuers, err := r.findIssuersBasedOnCA(caSecret)
			if err != nil {
				return reconcile.Result{}, err
			}

			var leafSecrets []*corev1.Secret

			for _, issuer := range issuers {
				leafSecrets, err = r.findLeafSecrets(issuer.Name, issuer.Namespace)
				if err != nil {
					log.Error(err, "Error reading the leaf certificates for issuer - requeue the request")
					return reconcile.Result{}, err
				}
			}

			for _, leafSecret := range leafSecrets {
				if err := r.client.Delete(context.TODO(), leafSecret); err != nil {
					if errors.IsNotFound(err) {
						continue
					}
					return reconcile.Result{}, err
				}
			}

			if err := r.updateLeafCerts(issuers); err != nil {
				return reconcile.Result{}, err
			}

		}

		reqLogger.Info("Created v1 Certificate")
	} else {
		// should be safe to assume that NotAfter exists at this point since if statement would have executed if it was empty
		t := time.Until(getExpiration(*instance.Status.NotAfter))
		reqLogger.Info("Not creating v1 Certificate because existing certificate secret still valid", "Requeuing in: %v", t)
		return reconcile.Result{RequeueAfter: t}, nil
	}

	return reconcile.Result{}, nil
}

// isExpired Determines if v1alpha1 Certificate is expired or not based on three
// conditions:
// 1. existence of NotAfter status
// 2. existence of certificate secret
// 3. is current date after expiration date
// TODO: could optionally inspect the secret to check if NotAfter date matches with Certificate status
func isExpired(c *certmanagerv1alpha1.Certificate, s *corev1.Secret) bool {
	if c.Status.NotAfter == nil {
		return true
	}
	if s == nil {
		return true
	}
	return time.Now().After(getExpiration(*c.Status.NotAfter))
}

// getExpiration Gets the time when Certificate would have been renewed by
// cert-manager controller. Subtracting one day to provide a buffer time
func getExpiration(notAfter metav1.Time) time.Time {
	return notAfter.Add(-time.Hour * 24)
}

// purgeOldV1 deletes all v1 Certificates generated by the operator before v1.x
// operand was deployed, i.e. before operator v3.14.0. New conversion logic
// will only conditionally create v1 Certificates, so previous ones must be
// deleted
func (r *ReconcileCertificate) purgeOldV1() error {
	oldV1List := &certmanagerv1.CertificateList{}
	if err := r.client.List(context.TODO(), oldV1List); err != nil {
		return err
	}
	for _, v := range oldV1List.Items {
		if v.Annotations[resources.OperatorGeneratedAnno] == t && v.Labels[resources.ProperV1Label] != t {
			if err := r.client.Delete(context.TODO(), &v); err != nil {
				return err
			}
		}
	}
	return nil
}

// getSecret finds corresponding secret of the certificate
func (r *ReconcileCertificate) getSecret(cert *certmanagerv1alpha1.Certificate) (*corev1.Secret, error) {
	secretName := cert.Spec.SecretName
	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: cert.Namespace}, secret)

	return secret, err
}

// findIssuersBasedOnCA finds issuers that are based on the given CA secret
func (r *ReconcileCertificate) findIssuersBasedOnCA(caSecret *corev1.Secret) ([]certmanagerv1alpha1.Issuer, error) {

	var issuers []certmanagerv1alpha1.Issuer

	issuerList := &certmanagerv1alpha1.IssuerList{}
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

// findLeafSecrets finds issuers that are based on the given CA secret
func (r *ReconcileCertificate) findLeafSecrets(issuedBy string, namespace string) ([]*corev1.Secret, error) {

	var leafSecrets []*corev1.Secret

	certList := &certmanagerv1alpha1.CertificateList{}
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

// updateLeafCerts adds a label to v1alpha1 leaf Certificate, so that it will be
// converted to v1 Certificate. The secret for the leaf certificate must be
// deleted beforehand.
func (r *ReconcileCertificate) updateLeafCerts(issuers []certmanagerv1alpha1.Issuer) error {
	for _, i := range issuers {
		certList := &certmanagerv1alpha1.CertificateList{}
		if err := r.client.List(context.TODO(), certList, &client.ListOptions{Namespace: i.Namespace}); err != nil {
			return err
		}

		for _, c := range certList.Items {
			if c.Spec.IssuerRef.Name == i.Name {
				if c.Labels == nil {
					c.Labels = make(map[string]string)
				}
				c.Labels["ibm-cert-manager-operator/conversion-leaf-refresh"] = "true"
				if err := r.client.Update(context.TODO(), &c); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *ReconcileCertificate) updateWebhooks(s string) error {
	mwebhooks := &admRegv1.MutatingWebhookConfigurationList{}
	if err := r.client.List(context.TODO(), mwebhooks); err != nil {
		return err
	}
	for _, w := range mwebhooks.Items {
		if w.Annotations != nil {
			if w.Annotations["certmanager.k8s.io/inject-ca-from"] == s {
				w.Annotations["cert-manager.io/inject-ca-from"] = s
				if err := r.client.Update(context.TODO(), &w); err != nil {
					return err
				}
			}
		}
	}

	vwebhooks := &admRegv1.ValidatingWebhookConfigurationList{}
	if err := r.client.List(context.TODO(), vwebhooks); err != nil {
		return err
	}
	for _, w := range vwebhooks.Items {
		if w.Annotations != nil {
			if w.Annotations["certmanager.k8s.io/inject-ca-from"] == s {
				w.Annotations["cert-manager.io/inject-ca-from"] = s
				if err := r.client.Update(context.TODO(), &w); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
