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

package certmanager

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-cert-manager-operator/pkg/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Returns true if no errors in deploy logic
func certManagerDeploy(instance *operatorv1alpha1.CertManager, client client.Client, kubeclient kubernetes.Interface, scheme *runtime.Scheme) error {
	return deployLogic(instance, client, kubeclient, scheme, res.ControllerDeployment, res.CertManagerControllerName, res.ControllerImageName, res.ControllerLabels)
}

func cainjectorDeploy(instance *operatorv1alpha1.CertManager, client client.Client, kubeclient kubernetes.Interface, scheme *runtime.Scheme) error {
	return deployLogic(instance, client, kubeclient, scheme, res.CainjectorDeployment, res.CertManagerCainjectorName, res.CainjectorImageName, res.CainjectorLabels)
}

func webhookDeploy(instance *operatorv1alpha1.CertManager, client client.Client, kubeclient kubernetes.Interface, scheme *runtime.Scheme) error {
	return deployLogic(instance, client, kubeclient, scheme, res.WebhookDeployment, res.CertManagerWebhookName, res.WebhookImageName, res.WebhookLabels)
}

func configmapWatcherDeploy(instance *operatorv1alpha1.CertManager, client client.Client, kubeclient kubernetes.Interface, scheme *runtime.Scheme) error {
	return deployLogic(instance, client, kubeclient, scheme, res.ConfigmapWatcherDeployment, res.ConfigmapWatcherName, res.ConfigmapWatcherImageName, res.ConfigmapWatcherLabels)
}

func deployLogic(instance *operatorv1alpha1.CertManager, client client.Client, kubeclient kubernetes.Interface, scheme *runtime.Scheme, deployTemplate *appsv1.Deployment, name, imageName, labels string) error {
	similarDeploys := deployFinder(kubeclient, labels, imageName)
	deployment := setupDeploy(instance, deployTemplate)
	var existingDeploy appsv1.Deployment
	create := true

	log.V(2).Info("Working on deploy logic", "deployment name", name)
	log.V(3).Info("Length of similar deployments found", "len", len(similarDeploys))
	log.V(4).Info("The similar deploys", "all of them", fmt.Sprintf("%v", similarDeploys))

	for _, deploy := range similarDeploys {
		if !(deploy.Name == name && deploy.Namespace == res.DeployNamespace) {
			// If there's more than one, and it's not the correct one, return an error with a warning
			errMsg := fmt.Sprintf("The service %s is already deployed as %s/%s. Please remove it if you want this version of %s to be deployed.",
				name, deploy.Namespace, deploy.Name, name)
			log.V(4).Info(errMsg)
			err := errors.New(errMsg)
			return err
		}
		// Otherwise one exists so we update it
		log.V(3).Info("Create false, this matches name and namespace", "deploy name", deploy.Name, "name", name)
		create = false
		existingDeploy = deploy
	}

	if err := controllerutil.SetControllerReference(instance, &deployment, scheme); err != nil {
		return err
	}
	if create {
		if err := client.Create(context.TODO(), &deployment); err != nil {
			return err
		}
	} else {
		if !equalDeploys(deployment, existingDeploy) {
			// Update
			log.V(2).Info("Updating deployment")
			if err := client.Update(context.Background(), &deployment); err != nil {
				return err
			}
		} else {
			log.V(3).Info("Deploys are equal, no changes needed")
		}
	}
	log.V(2).Info("Finished working on deploy logic", "deployment name", name)
	return nil
}

// Configure deployment options
// Args:deploy
//     instance - The CR instance of CertManager
//     deploy - The base deployment object - template contains most of the defaults/constants for the deployment
func setupDeploy(instance *operatorv1alpha1.CertManager, deploy *appsv1.Deployment) appsv1.Deployment {
	// First copy the deploy template into a deployment object

	returningDeploy := *deploy

	switch deploy.Name {
	case res.CertManagerControllerName:
		var acmesolver = res.AcmeSolverArg
		if instance.Spec.ImageRegistry != "" { // Set the image registry to the one specified
			// Assume image registry doesn't have forward slash at the end
			returningDeploy.Spec.Template.Spec.Containers[0].Image = instance.Spec.ImageRegistry + "/" + res.ControllerImageName + ":" + res.ControllerImageVersion
			acmesolver = "--acme-http01-solver-image=" + instance.Spec.ImageRegistry + "/" + res.AcmesolverImageName + ":" + res.ControllerImageVersion
		}
		var resourceNS = res.ResourceNS
		if instance.Spec.ResourceNS != "" {
			resourceNS = "--cluster-resource-namespace=" + instance.Spec.ResourceNS
		}
		var args []string
		copy(res.DefaultArgs, args)
		args = append(args, acmesolver, resourceNS)
		returningDeploy.Spec.Template.Spec.Containers[0].Args = args
		log.V(3).Info("The args", "args", deploy.Spec.Template.Spec.Containers[0].Args)
	case res.CertManagerCainjectorName:
		if instance.Spec.ImageRegistry != "" { // Set the image registry to the one specified
			// Assume image registry doesn't have forward slash at the end
			returningDeploy.Spec.Template.Spec.Containers[0].Image = instance.Spec.ImageRegistry + "/" + res.CainjectorImageName + ":" + res.ControllerImageVersion
		}
	case res.CertManagerWebhookName:
		if instance.Spec.ImageRegistry != "" { // Set the image registry to the one specified
			// Assume image registry doesn't have forward slash at the end
			returningDeploy.Spec.Template.Spec.Containers[0].Image = instance.Spec.ImageRegistry + "/" + res.WebhookImageName + ":" + res.WebhookImageVersion
			returningDeploy.Spec.Template.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem = &res.FalseVar
		}
	case res.ConfigmapWatcherName:
		if instance.Spec.ImageRegistry != "" { // Set the image registry to the one specified
			// Assume image registry doesn't have forward slash at the end
			returningDeploy.Spec.Template.Spec.Containers[0].Image = instance.Spec.ImageRegistry + "/" + res.ConfigmapWatcherImageName + ":" + res.ConfigmapWatcherVersion
		}
	}

	if instance.Spec.ImagePostFix != "" {
		returningDeploy.Spec.Template.Spec.Containers[0].Image += instance.Spec.ImagePostFix
	}

	log.V(2).Info("Resulting image registry", "full name", returningDeploy.Spec.Template.Spec.Containers[0].Image)
	log.V(3).Info("Resulting deployment to be created", "spec", fmt.Sprintf("%v", returningDeploy))
	return returningDeploy
}

func removeDeploy(client kubernetes.Interface, name, namespace string) error {
	if err := client.AppsV1().Deployments(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		log.V(1).Info("Error removing deployment", "name", name, "namespace", namespace, "error message", err)
		return err
	}
	log.V(2).Info("Deployment removed", "deploy name", name, "deploy namespace", namespace)
	return nil
}

func deployFinder(client kubernetes.Interface, labels, name string) []appsv1.Deployment {
	log.V(2).Info("Finding preexisting deployments", "deployment name", name)
	listOpt := metav1.ListOptions{LabelSelector: labels}
	var allDeploys []appsv1.Deployment
	var allDeploysMap = make(map[string]appsv1.Deployment)
	deployList, err := client.AppsV1().Deployments("").List(listOpt)
	// Find deployment by its labels
	if err != nil {
		log.Error(err, "Error retrieving deployments by label")
	} else {
		for _, deploy := range deployList.Items {
			log.V(3).Info("Found deployment by labels",
				"name", deploy.ObjectMeta.Name, "namespace",
				deploy.ObjectMeta.Namespace, "labels", fmt.Sprintf("%v", deploy.ObjectMeta.Labels))
			ns := deploy.Name + "/" + deploy.Namespace
			if _, ok := allDeploysMap[ns]; !ok {
				log.V(4).Info("adding to map", "name", deploy.Name, "namespace", deploy.Namespace)
				allDeploysMap[ns] = deploy
			}
		}
	}

	// Find deployment by querying for all and checking the image
	listOpt = metav1.ListOptions{}
	deployList, err = client.AppsV1().Deployments("").List(listOpt)
	// Check all the deployments in namespace to see if cert-manager is already installed as a different name
	if err != nil {
		log.Error(err, "Error retrieving deployments")
	} else {
		for _, deploy := range deployList.Items {
			if strings.Contains(deploy.Spec.Template.Spec.Containers[0].Image, name) { // Deploys the same image
				log.V(3).Info("Found deployment by image name",
					"name", deploy.ObjectMeta.Name, "namespace", deploy.ObjectMeta.Namespace,
					"image name", deploy.Spec.Template.Spec.Containers[0].Image)
				ns := deploy.Name + "/" + deploy.Namespace
				if _, ok := allDeploysMap[ns]; !ok {
					log.V(4).Info("adding to map")
					allDeploysMap[ns] = deploy
				}
			}
		}
	}
	for _, v := range allDeploysMap {
		log.V(4).Info("Appending deploy to slice", "name", v.Name, "namespace", v.Namespace)
		allDeploys = append(allDeploys, v)
	}
	return allDeploys
}

// Deep comparison between the two deployments passed in
// Checks labels, replicas, pod template labels, pull secrets, service account names,
// volumes, liveness, readiness, image name, args, env, and security contexts (pod & container)
// of both deployments. If there are any discrepencies between them, this returns false. Returns
// true otherwise
func equalDeploys(first, second appsv1.Deployment) bool {
	statusLog := log.V(1)
	if !reflect.DeepEqual(first.ObjectMeta.Labels, second.ObjectMeta.Labels) {
		statusLog.Info("Labels not equal",
			"first", fmt.Sprintf("%v", first.ObjectMeta.Labels),
			"second", fmt.Sprintf("%v", second.ObjectMeta.Labels))
		return false
	}

	if !reflect.DeepEqual(first.Spec.Replicas, second.Spec.Replicas) {
		statusLog.Info("Replicas not equal", "first", first.Spec.Replicas, "second", second.Spec.Replicas)
		return false
	}

	firstPodTemplate := first.Spec.Template
	secondPodTemplate := second.Spec.Template
	if !reflect.DeepEqual(firstPodTemplate.ObjectMeta.Labels, secondPodTemplate.ObjectMeta.Labels) {
		statusLog.Info("Pod labels not equal",
			"first", fmt.Sprintf("%v", firstPodTemplate.ObjectMeta.Labels),
			"second", fmt.Sprintf("%v", secondPodTemplate.ObjectMeta.Labels))
		return false
	}

	if !reflect.DeepEqual(firstPodTemplate.Spec.ImagePullSecrets, secondPodTemplate.Spec.ImagePullSecrets) {
		statusLog.Info("Image pull secrets not equal",
			"first", fmt.Sprintf("%v", firstPodTemplate.Spec.ImagePullSecrets),
			"second", fmt.Sprintf("%v", secondPodTemplate.Spec.ImagePullSecrets))
		return false
	}

	if !reflect.DeepEqual(firstPodTemplate.Spec.ServiceAccountName, secondPodTemplate.Spec.ServiceAccountName) {
		statusLog.Info("Service account names not equal",
			"first", firstPodTemplate.Spec.ServiceAccountName,
			"second", secondPodTemplate.Spec.ServiceAccountName)
		return false
	}

	if !reflect.DeepEqual(firstPodTemplate.Spec.SecurityContext, secondPodTemplate.Spec.SecurityContext) {
		statusLog.Info("Security context not equal",
			"first", fmt.Sprintf("%v", firstPodTemplate.Spec.SecurityContext),
			"second", fmt.Sprintf("%v", secondPodTemplate.Spec.SecurityContext))
		return false
	}
	fVol := firstPodTemplate.Spec.Volumes
	sVol := secondPodTemplate.Spec.Volumes
	if reflect.DeepEqual(len(fVol), len(sVol)) {
		if len(fVol) > 0 {
			for i := range fVol {
				if !reflect.DeepEqual(fVol[i].Name, sVol[i].Name) {
					statusLog.Info("Pod volume names not equal", "volume num", i,
						"first", fVol[i].Name, "second", sVol[i].Name)
					return false
				}
				if fVol[i].VolumeSource.Secret != nil && sVol[i].VolumeSource.Secret != nil {
					if !reflect.DeepEqual(fVol[i].VolumeSource.Secret.SecretName, sVol[i].VolumeSource.Secret.SecretName) {
						statusLog.Info("Volume source secret name not equal", "volume num", i,
							"first", fVol[i].VolumeSource.Secret.SecretName, "second", sVol[i].VolumeSource.Secret.SecretName)
						return false
					}
				} else if !(fVol[i].VolumeSource.Secret == nil && sVol[i].VolumeSource.Secret == nil) {
					statusLog.Info("One of the volume sources secrets is nil")
					return false
				}
			}
		}
	} else {
		statusLog.Info("Volume lengths not equal")
		return false
	}

	// Container level checks
	firstContainers := firstPodTemplate.Spec.Containers
	secondContainers := secondPodTemplate.Spec.Containers
	if len(firstContainers) != len(secondContainers) {
		statusLog.Info("Number of containers not equal",
			"first", len(firstContainers), "second", len(secondContainers))
		return false
	}

	fContainer := firstContainers[0]
	sContainer := secondContainers[0]
	if !reflect.DeepEqual(fContainer.Name, sContainer.Name) {
		statusLog.Info("Container names not equal", "first", fContainer.Name, "second", sContainer.Name)
		return false
	}

	if !reflect.DeepEqual(fContainer.Image, sContainer.Image) {
		statusLog.Info("Container images not equal", "first", fContainer.Image, "second", sContainer.Image)
		return false
	}

	if !reflect.DeepEqual(fContainer.ImagePullPolicy, sContainer.ImagePullPolicy) {
		statusLog.Info("Image pull policies not equal",
			"first", fContainer.ImagePullPolicy, "second", sContainer.ImagePullPolicy)
		return false
	}

	if fContainer.Args != nil && sContainer.Args != nil {
		if !reflect.DeepEqual(len(fContainer.Args), len(sContainer.Args)) {
			statusLog.Info("Args length not equal",
				"first", len(fContainer.Args), "second", len(sContainer.Args))
			return false
		}
		if !reflect.DeepEqual(fContainer.Args, sContainer.Args) {
			statusLog.Info("Args not equal",
				"first", fmt.Sprintf("%v", fContainer.Args), "second", fmt.Sprintf("%v", sContainer.Args))
			return false
		}
	} else if !(fContainer.Args == nil && sContainer.Args == nil) {
		statusLog.Info("One of the args is nil",
			"first", fmt.Sprintf("%v", fContainer.Args), "second", fmt.Sprintf("%v", sContainer.Args))
		return false
	}

	fLive := fContainer.LivenessProbe
	sLive := sContainer.LivenessProbe

	if fLive != nil && sLive != nil {
		if !reflect.DeepEqual(fLive.Handler.Exec.Command, sLive.Handler.Exec.Command) {
			statusLog.Info("Exec command in liveness probes not equal",
				"first", fLive.Handler.Exec.Command, "second", sLive.Handler.Exec.Command)
			return false
		}

		if !reflect.DeepEqual(fLive.InitialDelaySeconds, sLive.InitialDelaySeconds) {
			statusLog.Info("Initial delay seconds in liveness probes not equal",
				"first", fLive.InitialDelaySeconds, "second", sLive.InitialDelaySeconds)
			return false
		}

		if !reflect.DeepEqual(fLive.TimeoutSeconds, sLive.TimeoutSeconds) {
			statusLog.Info("Timeout seconds in liveness probes not equal",
				"first", fLive.TimeoutSeconds, "second", sLive.TimeoutSeconds)
			return false
		}
	} else if !(fLive == nil && sLive == nil) {
		statusLog.Info("One liveness probe is nil",
			"first", fmt.Sprintf("%v", fLive), "second", fmt.Sprintf("%v", sLive))
		return false
	}

	fReady := fContainer.ReadinessProbe
	sReady := sContainer.ReadinessProbe
	if fReady != nil && sReady != nil {
		if !reflect.DeepEqual(fReady.Handler.Exec.Command, sReady.Handler.Exec.Command) {
			statusLog.Info("Exec command in readiness probes not equal",
				"first", fReady.Handler.Exec.Command, "second", sReady.Handler.Exec.Command)
			return false
		}

		if !reflect.DeepEqual(fReady.InitialDelaySeconds, sReady.InitialDelaySeconds) {
			statusLog.Info("Initial delay seconds in readiness probes not equal",
				"first", fReady.InitialDelaySeconds, "second", sReady.InitialDelaySeconds)
			return false
		}

		if !reflect.DeepEqual(fReady.TimeoutSeconds, sReady.TimeoutSeconds) {
			statusLog.Info("Timeout seconds in readiness probes not equal",
				"first", fReady.TimeoutSeconds, "second", sReady.TimeoutSeconds)
			return false
		}
	} else if !(fReady == nil && sReady == nil) {
		statusLog.Info("One of the readiness probes is nil",
			"first", fmt.Sprintf("%v", fReady), "second", fmt.Sprintf("%v", sReady))
		return false
	}

	fSecCont := fContainer.SecurityContext
	sSecCont := sContainer.SecurityContext

	if fSecCont != nil && sSecCont != nil {
		if fSecCont.RunAsNonRoot != nil && sSecCont.RunAsNonRoot != nil {
			if !reflect.DeepEqual(fSecCont.RunAsNonRoot, sSecCont.RunAsNonRoot) {
				statusLog.Info("Container security context run as non root not equal",
					"first", fSecCont.RunAsNonRoot, "second", sSecCont.RunAsNonRoot)
				return false
			}
		} else if !(fSecCont.RunAsNonRoot == nil && sSecCont.RunAsNonRoot == nil) {
			statusLog.Info("One security context run as non root is nil")
			return false
		}

		if fSecCont.RunAsUser != nil && sSecCont.RunAsUser != nil {
			if !reflect.DeepEqual(fSecCont.RunAsUser, sSecCont.RunAsUser) {
				statusLog.Info("Container security context run as user not equal",
					"first", fSecCont.RunAsUser, "second", sSecCont.RunAsUser)
				return false
			}
		} else if !(fSecCont.RunAsUser == nil && sSecCont.RunAsUser == nil) {
			statusLog.Info("One security context run as user is nil")
			return false
		}

		if fSecCont.AllowPrivilegeEscalation != nil && sSecCont.AllowPrivilegeEscalation != nil {
			if !reflect.DeepEqual(fSecCont.AllowPrivilegeEscalation, sSecCont.AllowPrivilegeEscalation) {
				statusLog.Info("Container security context AllowPrivilegeEscalation not equal",
					"first", fSecCont.AllowPrivilegeEscalation, "second", sSecCont.AllowPrivilegeEscalation)
				return false
			}
		} else if !(fSecCont.AllowPrivilegeEscalation == nil && sSecCont.AllowPrivilegeEscalation == nil) {
			statusLog.Info("One security context AllowPrivilegeEscalation is nil")
			return false
		}

		if fSecCont.ReadOnlyRootFilesystem != nil && sSecCont.ReadOnlyRootFilesystem != nil {
			if !reflect.DeepEqual(fSecCont.ReadOnlyRootFilesystem, sSecCont.ReadOnlyRootFilesystem) {
				statusLog.Info("Container security context ReadOnlyRootFilesystem not equal",
					"first", fSecCont.ReadOnlyRootFilesystem, "second", sSecCont.ReadOnlyRootFilesystem)
				return false
			}
		} else if !(fSecCont.ReadOnlyRootFilesystem == nil && sSecCont.ReadOnlyRootFilesystem == nil) {
			statusLog.Info("One security context ReadOnlyRootFilesystem is nil")
			return false
		}

		if fSecCont.Privileged != nil && sSecCont.Privileged != nil {
			if !reflect.DeepEqual(fSecCont.Privileged, sSecCont.Privileged) {
				statusLog.Info("Container security context Privileged not equal",
					"first", fSecCont.Privileged, "second", sSecCont.Privileged)
				return false
			}
		} else if !(fSecCont.Privileged == nil && sSecCont.Privileged == nil) {
			statusLog.Info("One security context Privileged is nil")
			return false
		}

		if fSecCont.Capabilities != nil && sSecCont.Capabilities != nil {
			if !reflect.DeepEqual(fSecCont.Capabilities, sSecCont.Capabilities) {
				statusLog.Info("Container security context Capabilities not equal",
					"first", fSecCont.Capabilities, "second", sSecCont.Capabilities)
				return false
			}
		} else if !(fSecCont.Capabilities == nil && sSecCont.Capabilities == nil) {
			statusLog.Info("One security context Capabilities is nil")
			return false
		}
	} else if !(fSecCont == nil && sSecCont == nil) {
		statusLog.Info("One security context is nil")
		return false
	}

	fRes := fContainer.Resources
	sRes := sContainer.Resources

	if !reflect.DeepEqual(fRes.Limits[corev1.ResourceCPU], sRes.Limits[corev1.ResourceCPU]) {
		statusLog.Info("Resource limit cpu not equal",
			"first", fRes.Limits[corev1.ResourceCPU], "second", sRes.Limits[corev1.ResourceCPU])
		return false
	}

	if !reflect.DeepEqual(fRes.Limits[corev1.ResourceMemory], sRes.Limits[corev1.ResourceMemory]) {
		statusLog.Info("Resource limit memory not equal",
			"first", fRes.Limits[corev1.ResourceMemory], "second", sRes.Limits[corev1.ResourceMemory])
		return false
	}

	if !reflect.DeepEqual(fRes.Requests[corev1.ResourceCPU], sRes.Requests[corev1.ResourceCPU]) {
		statusLog.Info("Resource requests cpu not equal",
			"first", fRes.Requests[corev1.ResourceCPU], "second", sRes.Requests[corev1.ResourceCPU])
		return false
	}

	if !reflect.DeepEqual(fRes.Requests[corev1.ResourceMemory], sRes.Requests[corev1.ResourceMemory]) {
		statusLog.Info("Resource requests memory not equal",
			"first", fRes.Requests[corev1.ResourceMemory], "second", sRes.Requests[corev1.ResourceMemory])
		return false
	}

	fEnv := fContainer.Env
	sEnv := sContainer.Env
	if !reflect.DeepEqual(len(fEnv), len(sEnv)) {
		statusLog.Info("Environment var length not equal")
		return false
	} else if len(fEnv) > 0 {
		for i := range fEnv {
			if !reflect.DeepEqual(fEnv[i].Name, sEnv[i].Name) {
				statusLog.Info("Container number", "first", i)
				statusLog.Info("Environment names not equal", "first", fEnv[i].Name, "second", sEnv[i].Name)
				return false
			}
			if !reflect.DeepEqual(fEnv[i].Value, sEnv[i].Value) {
				statusLog.Info("Container number", "first", i)
				statusLog.Info("Environment values not equal", "first", fEnv[i].Value, "second", sEnv[i].Value)
				return false
			}
			if fEnv[i].ValueFrom != nil && sEnv[i].ValueFrom != nil {
				fFieldRef := fEnv[i].ValueFrom.FieldRef
				sFieldRef := sEnv[i].ValueFrom.FieldRef
				if fFieldRef != nil && sFieldRef != nil {
					if !reflect.DeepEqual(fEnv[i].ValueFrom.FieldRef.FieldPath, sEnv[i].ValueFrom.FieldRef.FieldPath) {
						statusLog.Info("Field path in env not equal",
							"first", fEnv[i].ValueFrom.FieldRef.FieldPath, "second", sEnv[i].ValueFrom.FieldRef.FieldPath)
						return false
					}
				} else if !(fFieldRef == nil && sFieldRef == nil) {
					statusLog.Info("Container number", "first", i)
					statusLog.Info("One of the env's field ref is nil")
					return false
				}

			} else if !(fEnv[i].ValueFrom == nil && sEnv[i].ValueFrom == nil) {
				statusLog.Info("Container number", "first", i)
				statusLog.Info("One of the env's value from is nil")
				return false
			}
		}
	}
	fVolMnt := fContainer.VolumeMounts
	sVolMnt := sContainer.VolumeMounts
	if reflect.DeepEqual(len(fVolMnt), len(sVolMnt)) {
		if len(fVolMnt) > 0 {
			for i := range fVolMnt {
				if !reflect.DeepEqual(fVolMnt[i], sVolMnt[i]) {
					statusLog.Info("Volume mounts not equal", "num", i,
						"first", fmt.Sprintf("%v", fVolMnt[i]), "second", fmt.Sprintf("%v", sVolMnt[i]))
					return false
				}
			}
		}
	} else {
		statusLog.Info("Volume mount lengths not equal")
		return false
	}

	log.V(2).Info("Finished checking for differences between the deployments and found none.", "deployment name", first.Name)
	return true
}
