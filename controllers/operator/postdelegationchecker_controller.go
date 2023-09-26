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
	"os"
	"strings"

	"github.com/go-logr/logr"
	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/operator/v1alpha1"
	olmv1 "github.com/operator-framework/api/pkg/operators/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const CERTIFICATE_API = "Certificate.v1.cert-manager.io"
const POST_DELEGATION_CONFIG = "disablePostDelegation"
const OPERAND_TOGGLE = "deployCSCertManagerOperands"

// PostDelegationCheckerReconciler reconciles a PostDelegationChecker object
type PostDelegationCheckerReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operators.coreos.com,resources=operatorgroups,verbs=get;list;watch

// Reconcile is run whenever an OperatorGroup is created or updated on the cluster.
// If the OperatorGroup's annotations contains 'Certificate.cert-manager.io`, then
// the foundational services' cert-manager operands are deleted to delegate cert-manager
// responsibility to the newly installed cert-manager on the cluster
func (r *PostDelegationCheckerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	logger.Info("Reconciling OperatorGroup for delegation check")

	configmap := &corev1.ConfigMap{}
	ns, _ := os.LookupEnv("WATCH_NAMESPACE")

	if err := r.Client.Get(ctx, types.NamespacedName{
		Name:      "ibm-cpp-config",
		Namespace: ns,
	}, configmap); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		logger.V(2).Info("Configmap ibm-cpp-config not found, skipping check")
	}
	if v, ok := configmap.Data[POST_DELEGATION_CONFIG]; ok {
		if v == "true" {
			logger.Info("Post delegation check disabled, skipping")
			return ctrl.Result{}, nil
		}
	}

	instance := &olmv1.OperatorGroup{}

	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			logger.V(2).Info("OperatorGroup not found, probably deleted, not requeuing")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if _, exists := instance.Annotations["olm.providedAPIs"]; !exists {
		logger.Info("OperatorGroup has no olm.ProvidedAPIs annotation, not requeuing")
		return ctrl.Result{}, nil
	}

	if a := instance.Annotations["olm.providedAPIs"]; a == "" {
		logger.Info("OperatorGroup has empty olm.ProvidedAPIs annotation, not requeuing")
		return ctrl.Result{}, nil
	}

	if exists := strings.Index(instance.Annotations["olm.providedAPIs"], CERTIFICATE_API); exists == -1 {
		logger.Info("OperatorGroup olm.ProvidedAPIs does not include " + CERTIFICATE_API)
		return ctrl.Result{}, nil
	}

	logger.Info("Removing operands because OperatorGroup olm.ProvidedAPIs contains " + CERTIFICATE_API)

	return r.deleteOperand(ctx, logger)
}

// deleteOperand deletes the CertManager.operator.ibm.com instance named default
// and configures ibm-cpp-config so that the operands to not come back. The configmap
// configuration is to avoid situations with foundational services' cert-manager
// returning because of ODLM creating the CR before the other cert-manager has
// finished installation.
func (r *PostDelegationCheckerReconciler) deleteOperand(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {

	logger.Info("Configuring ibm-cpp-config to permanently disable operands")

	configmap := &corev1.ConfigMap{}
	ns, _ := os.LookupEnv("WATCH_NAMESPACE")

	configmap.Name = "ibm-cpp-config"
	configmap.Namespace = ns

	if err := r.Client.Create(ctx, configmap); err != nil {
		if !errors.IsAlreadyExists(err) {
			return ctrl.Result{}, err
		}
	}

	if configmap.Data == nil {
		configmap.Data = map[string]string{}
	}
	configmap.Data[OPERAND_TOGGLE] = "false"

	if err := r.Client.Update(ctx, configmap); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Successfully configured ibm-cpp-config")

	logger.Info("Deleting default " + CERTIFICATE_API + " instance")

	certManagerObject := &operatorv1alpha1.CertManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
	}

	if err := r.Client.Delete(ctx, certManagerObject); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(CERTIFICATE_API + " instance not found, skip deletion")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("Successfully deleted" + CERTIFICATE_API + " instance to remove operands")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostDelegationCheckerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("postdelegationchecker_controller").
		For(&olmv1.OperatorGroup{}).
		Complete(r)
}
