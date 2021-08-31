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

package podrefresh

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1"
)

var log = logf.Log.WithName("controller_podrefresh")
var (
	// TODO support cert-manager.io
	expirationLabel     = "cert-manager.io/expiration"
	restartLabel        = "certmanager.k8s.io/time-restarted"
	noRestartAnnotation = "certmanager.k8s.io/disable-auto-restart"
	t                   = "true"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new podrefresh Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &Reconcilepodrefresh{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("podrefresh-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Certificates in the cluster
	err = c.Watch(&source.Kind{Type: &certmanagerv1.Certificate{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that Reconcilepodrefresh implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconcilepodrefresh{}

// Reconcilepodrefresh reconciles a podrefresh object
type Reconcilepodrefresh struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Certificate object and makes changes based on the state read
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconcilepodrefresh) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling podrefresh")

	// Get the certificate that invoked reconciliation is a CA in the listOfCAs

	cert := &certmanagerv1.Certificate{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cert)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if len(cert.Status.Conditions) > 0 && cert.Status.NotAfter != nil {
		if err := r.restart(cert.Spec.SecretName, cert.Name, cert.Status.NotAfter.Format("2006-1-2.1504")); err != nil {
			reqLogger.Error(err, "Failed to fresh pod")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// pod refresh is enabled. It will edit the deployments, statefulsets, and daemonsets
// that use the secret being updated, which will trigger the pod to be restarted.
func (r *Reconcilepodrefresh) restart(secret, cert string, expiration string) error {
	timeNow := time.Now().Format("2006-1-2.1504")
	deployments := &appsv1.DeploymentList{}
	if err := r.client.List(context.TODO(), deployments); err != nil {
		return fmt.Errorf("error getting deployments: %v", err)
	}
	deploymentsToUpdate, err := r.getDeploymentsNeedUpdate(secret, expiration)
	if err != nil {
		return err
	}

	if err := r.updateDeploymentAnnotations(deploymentsToUpdate, cert, secret, timeNow, expiration); err != nil {
		return err
	}

	statefulsetsToUpdate, err := r.getStsNeedUpdate(secret, expiration)
	if err != nil {
		return err
	}
	if err := r.updateStsAnnotations(statefulsetsToUpdate, cert, secret, timeNow, expiration); err != nil {
		return err
	}

	daemonsetsToUpdate, err := r.getDaemonSetNeedUpdate(secret, expiration)
	if err != nil {
		return err
	}
	if err := r.updateDaemonSetAnnotations(daemonsetsToUpdate, cert, secret, timeNow, expiration); err != nil {
		return err
	}

	return nil
}

func (r *Reconcilepodrefresh) getDeploymentsNeedUpdate(secret, expiration string) ([]appsv1.Deployment, error) {
	deploymentsToUpdate := make([]appsv1.Deployment, 0)
	deployments := &appsv1.DeploymentList{}
	if err := r.client.List(context.TODO(), deployments); err != nil {
		return deploymentsToUpdate, fmt.Errorf("error getting deployments: %v", err)
	}
NEXT_DEPLOYMENT:
	for _, deployment := range deployments.Items {
		if deployment.ObjectMeta.Labels != nil {
			if expiration == deployment.ObjectMeta.Labels[expirationLabel] {
				continue
			}
		}
		for _, container := range deployment.Spec.Template.Spec.Containers {
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secret && deployment.ObjectMeta.Annotations[noRestartAnnotation] != t {
					deploymentsToUpdate = append(deploymentsToUpdate, deployment)
					continue NEXT_DEPLOYMENT
				}
			}
		}
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.Secret != nil && volume.Secret.SecretName != "" && volume.Secret.SecretName == secret && deployment.ObjectMeta.Annotations[noRestartAnnotation] != t {
				deploymentsToUpdate = append(deploymentsToUpdate, deployment)
				continue NEXT_DEPLOYMENT
			}
			if volume.Projected != nil && volume.Projected.Sources != nil && deployment.ObjectMeta.Annotations[noRestartAnnotation] != t {
				for _, source := range volume.Projected.Sources {
					if source.Secret != nil && source.Secret.Name == secret {
						deploymentsToUpdate = append(deploymentsToUpdate, deployment)
						continue NEXT_DEPLOYMENT
					}
				}
			}
		}
	}
	return deploymentsToUpdate, nil
}

func (r *Reconcilepodrefresh) getStsNeedUpdate(secret, expiration string) ([]appsv1.StatefulSet, error) {
	statefulsetsToUpdate := make([]appsv1.StatefulSet, 0)
	statefulsets := &appsv1.StatefulSetList{}
	err := r.client.List(context.TODO(), statefulsets)
	if err != nil {
		return statefulsetsToUpdate, fmt.Errorf("error getting statefulsets: %v", err)
	}
NEXT_STATEFULSET:
	for _, statefulset := range statefulsets.Items {
		if statefulset.ObjectMeta.Labels != nil {
			if expiration == statefulset.ObjectMeta.Labels[expirationLabel] {
				continue
			}
		}
		for _, container := range statefulset.Spec.Template.Spec.Containers {
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secret && statefulset.ObjectMeta.Annotations[noRestartAnnotation] != t {
					statefulsetsToUpdate = append(statefulsetsToUpdate, statefulset)
					continue NEXT_STATEFULSET
				}
			}
		}
		for _, volume := range statefulset.Spec.Template.Spec.Volumes {
			if volume.Secret != nil && volume.Secret.SecretName != "" && volume.Secret.SecretName == secret && statefulset.ObjectMeta.Annotations[noRestartAnnotation] != t {
				statefulsetsToUpdate = append(statefulsetsToUpdate, statefulset)
				continue NEXT_STATEFULSET
			}
			if volume.Projected != nil && volume.Projected.Sources != nil && statefulset.ObjectMeta.Annotations[noRestartAnnotation] != t {
				for _, source := range volume.Projected.Sources {
					if source.Secret != nil && source.Secret.Name == secret {
						statefulsetsToUpdate = append(statefulsetsToUpdate, statefulset)
						continue NEXT_STATEFULSET
					}
				}
			}
		}
	}
	return statefulsetsToUpdate, nil
}

func (r *Reconcilepodrefresh) getDaemonSetNeedUpdate(secret, expiration string) ([]appsv1.DaemonSet, error) {
	daemonsetsToUpdate := make([]appsv1.DaemonSet, 0)
	daemonsets := &appsv1.DaemonSetList{}
	if err := r.client.List(context.TODO(), daemonsets); err != nil {
		return daemonsetsToUpdate, fmt.Errorf("error getting daemonsets: %v", err)
	}
NEXT_DAEMONSET:
	for _, daemonset := range daemonsets.Items {
		if daemonset.ObjectMeta.Labels != nil {
			if expiration == daemonset.ObjectMeta.Labels[expirationLabel] {
				continue
			}
		}
		for _, container := range daemonset.Spec.Template.Spec.Containers {
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == secret && daemonset.ObjectMeta.Annotations[noRestartAnnotation] != t {
					daemonsetsToUpdate = append(daemonsetsToUpdate, daemonset)
					continue NEXT_DAEMONSET
				}
			}
		}
		for _, volume := range daemonset.Spec.Template.Spec.Volumes {
			if volume.Secret != nil && volume.Secret.SecretName != "" && volume.Secret.SecretName == secret && daemonset.ObjectMeta.Annotations[noRestartAnnotation] != t {
				daemonsetsToUpdate = append(daemonsetsToUpdate, daemonset)
				continue NEXT_DAEMONSET
			}
			if volume.Projected != nil && volume.Projected.Sources != nil && daemonset.ObjectMeta.Annotations[noRestartAnnotation] != t {
				for _, source := range volume.Projected.Sources {
					if source.Secret != nil && source.Secret.Name == secret {
						daemonsetsToUpdate = append(daemonsetsToUpdate, daemonset)
						continue NEXT_DAEMONSET
					}
				}
			}
		}
	}
	return daemonsetsToUpdate, nil
}

func (r *Reconcilepodrefresh) updateDeploymentAnnotations(deploymentsToUpdate []appsv1.Deployment, cert, secret, timeNow, expiration string) error {
	for _, deployment := range deploymentsToUpdate {
		//in case of deployments not having labels section, create the label section
		if deployment.ObjectMeta.Labels == nil {
			deployment.ObjectMeta.Labels = make(map[string]string)
		}
		deployment.ObjectMeta.Labels[restartLabel] = timeNow
		deployment.Spec.Template.ObjectMeta.Labels[restartLabel] = timeNow
		deployment.ObjectMeta.Labels[expirationLabel] = expiration
		deployment.Spec.Template.ObjectMeta.Labels[expirationLabel] = expiration
		err := r.client.Update(context.TODO(), &deployment)
		if err != nil {
			return fmt.Errorf("error updating deployment: %v", err)
		}
		log.Info(timeNow, " Cert-Manager Restarting Resource:", "Certificate=", cert, "Secret=", secret, "Deployment=", deployment.ObjectMeta.Name)
	}
	return nil
}

func (r *Reconcilepodrefresh) updateStsAnnotations(statefulsetsToUpdate []appsv1.StatefulSet, cert, secret, timeNow, expiration string) error {
	for _, statefulset := range statefulsetsToUpdate {
		statefulset.ObjectMeta.Labels[restartLabel] = timeNow
		statefulset.Spec.Template.ObjectMeta.Labels[restartLabel] = timeNow
		statefulset.ObjectMeta.Labels[expirationLabel] = expiration
		statefulset.Spec.Template.ObjectMeta.Labels[expirationLabel] = expiration
		if err := r.client.Update(context.TODO(), &statefulset); err != nil {
			return fmt.Errorf("error updating statefulset: %v", err)
		}
		log.Info(timeNow, " Cert-Manager Restarting Resource:", "Certificate=", cert, "Secret=", secret, "StatefulSet=", statefulset.ObjectMeta.Name)
	}
	return nil
}

func (r *Reconcilepodrefresh) updateDaemonSetAnnotations(daemonsetsToUpdate []appsv1.DaemonSet, cert, secret, timeNow, expiration string) error {
	for _, daemonset := range daemonsetsToUpdate {
		daemonset.ObjectMeta.Labels[restartLabel] = timeNow
		daemonset.Spec.Template.ObjectMeta.Labels[restartLabel] = timeNow
		daemonset.ObjectMeta.Labels[expirationLabel] = expiration
		daemonset.Spec.Template.ObjectMeta.Labels[expirationLabel] = expiration
		if err := r.client.Update(context.TODO(), &daemonset); err != nil {
			return fmt.Errorf("error updating daemonset: %v", err)
		}
		log.Info(timeNow, " Cert-Manager Restarting Resource:", "Certificate=", cert, "Secret=", secret, "DaemonSet=", daemonset.ObjectMeta.Name)
	}
	return nil
}
