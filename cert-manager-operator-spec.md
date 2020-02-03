# Developing the Cert-Manager Operator 

## Overview

- Read [Operator Guidelines](https://github.ibm.com/IBMPrivateCloud/roadmap/blob/master/feature-specs/common-services/operator-guideline/operator-guideline-spec.md)
  to learn about the guidelines for Common Services operator.

- An operator can manage one or more controllers. The controller watches the resources for a particular CR (Custom Resource).

- All of the resources that used to be created via a helm chart will be created via a controller.

- Determine how many CRDs (Custom Resource Definition) are needed. Cert-manager will have 1 CRD:
    - CertManager - cert-manager-controller, cert-manager-webhook, cert-manager-cainjector

  
## Development

- These steps are based on [Operator Framework: Getting Started](https://github.com/operator-framework/getting-started#getting-started)
  and [Creating an App Operator](https://github.com/operator-framework/operator-sdk#create-and-deploy-an-app-operator).

- Repositories
  - https://github.com/IBM/ibm-cert-manager-operator
  - https://github.ibm.com/Crystal-Chun/ibm-cert-manager-operator

- Set the Go environment variables.

  `export GOPATH=/home/<username>/go`  
  `export GO111MODULE=on`  
  `export GOPRIVATE="github.ibm.com"`


- Create the operator skeleton.
  - `cd /home/ibmadmin/go/src/github.com/ibm`
  - `operator-sdk new ibm-cert-manager-operator --repo github.com/ibm/ibm-cert-manager-operator`
  - the main program for the operator, `cmd/manager/main.go`, initializes and runs the Manager
  - the Manager will automatically register the scheme for all custom resources defined under `pkg/apis/...`
    and run all controllers under `pkg/controller/...`
  - the Manager can restrict the namespace that all controllers will watch for resources

- Create the API definition ("Kind") which is used to create the CRD
  - `cd /home/ibmadmin/go/src/github.com/ibm/ibm-cert-manager-operator`
  - create `hack/boilerplate.go.txt`
	- contains copyright for generated code
  - `operator-sdk add api --api-version=operator.ibm.com/v1alpha1 --kind=CertManager`
	- generates `pkg/apis/operator/v1alpha1/<kind>_types.go`
	  - example: `pkg/apis/operator/v1alpha1/certmanager_types.go`
    - generates `deploy/crds/operator.ibm.com_<kind>s_crd.yaml`
      - example: `deploy/crds/operator.ibm.com_certmanagers_crd.yaml`
    - generates `deploy/crds/operator.ibm.com_v1alpha1_<kind>_cr.yaml`
      - example: `deploy/crds/operator.ibm.com_v1alpha1_certmanager_cr.yaml`
  - the operator can manage more than 1 Kind

- Edit `<kind>_types.go` and add the fields that will be exposed to the user. Then regenerate the CRD.
  - edit `<kind>_types.go` and add fields to the `<Kind>Spec` struct
  - `operator-sdk generate k8s`
	- updates `zz_generated.deepcopy.go`
  - "Operator Framework: Getting Started" says to run `operator-sdk generate openapi`. That command is deprecated, so run the next 2 commands instead.
    - `operator-sdk generate crds`
	  - updates `operator.ibm.com_certmanagers_crd.yaml`
    - `openapi-gen --logtostderr=true -o "" -i ./pkg/apis/operator/v1alpha1 -O zz_generated.openapi -p ./pkg/apis/operator/v1alpha1 -h hack/boilerplate.go.txt -r "-"`
      - creates `zz_generated.openapi.go`
      - if you need to build `openapi-gen`, follow these steps. The binary will be built in `$GOPATH/bin`.
        ```
        git clone https://github.com/kubernetes/kube-openapi.git
        cd kube-openapi
        go mod tidy
        go build -o ./bin/openapi-gen k8s.io/kube-openapi/cmd/openapi-gen
        ```
  - anytime you modify `<kind>_types.go`, run `generate k8s`, `generate crds`, and `openapi-gen` again to update the CRD and the generated code

- Create the controller. It will create resources like Deployments, DaemonSets, etc.
  - `operator-sdk add controller --api-version=operator.ibm.com/v1alpha1 --kind=CertManager`
  - there is 1 controller for each Kind/CRD
  - the controller will watch and reconcile the resources owned by the CR
  - for information about the Go types that implement Deployments, DaemonSets, etc, go to https://godoc.org/k8s.io/api/apps/v1
  - for information about the Go types that implement Pods, VolumeMounts, etc, go to https://godoc.org/k8s.io/api/core/v1
  - for information about the Go types that implement Ingress, etc, go to https://godoc.org/k8s.io/api/networking/v1beta1

## Testing
- Create the CRD. Do this one time before starting the operator.
  - `cd /home/ibmadmin/go/src/github.com/ibm/ibm-cert-manager-operator`
  - `oc login...`
  - `kubectl create -f deploy/crds/operator.ibm.com_certmanagers_crd.yaml`
  - `kubectl get crd | grep certmanager`
  - delete and create again if the CRD changes
    - `kubectl delete crd certmanagers.operator.ibm.com`

- Run the operator on a cluster
  - `cd /home/ibmadmin/go/src/github.com/ibm/ibm-cert-manager-operator`
  - Build your image & push to docker repo: `operator-sdk build hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com/crystal/operator/cert-manager-operator:1 && docker push hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com/crystal/operator/cert-manager-operator:1`
  - Log into your cluster `oc login ...`
  - Deploy your operator
    - Deploy resources it needs such as ClusterRole/ClusterRoleBinding `kubectl create -f deploy/clusterrole.yaml` `kubectl create -f deploy/clusterrolebinding.yaml` 
    - Edit the operator.yaml file to point to the image specified above
    - `kubectl create -f deploy/operator.yaml`


- Create a CR which is an instance of the CRD
  - edit `deploy/crds/operator.ibm.com_v1alpha1_certmanager_cr.yaml`
  - `kubectl create -f deploy/crds/operator.ibm.com_v1alpha1_certmanager_cr.yaml`

- Delete the CR and the associated resources that were created
  - `kubectl delete certmanager example-certmanager`
