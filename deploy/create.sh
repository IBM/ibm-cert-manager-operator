#!/bin/bash
echo "Creating image pull secret"
kubectl create secret docker-registry pull-secret-1 -n cert-manager --docker-server=hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com --docker-username=$1 --docker-password=$2
if [[ "$?" -ne 0 ]] ; then echo "Failed creating image pull secret" ; exit $? ; fi

echo "Creating CRDs"
kubectl create -f $(pwd)/crds/operator.ibm.com_certmanagers_crd.yaml
if [[ "$?" -ne 0 ]] ; then echo "Failed creating crds." ; exit $? ; fi

echo "Creating RBAC"
kubectl create -f $(pwd)/role.yaml && kubectl create -f $(pwd)/role_binding.yaml
if [[ "$?" -ne 0 ]] ; then echo "Failed creating RBAC." ; exit $? ; fi

echo "Creating service account"
kubectl create -f $(pwd)/service_account.yaml
if [[ "$?" -ne 0 ]] ; then echo "Failed creating service account." ; exit $? ; fi

echo "Deploying cert-manager-operator"
imageName=${3:-hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com/crystal/test/operator:41}
imageName=$(echo "$imageName" | sed 's/\//\\\//g')
sed -i "s/IMAGE_NAME/$imageName/" $(pwd)/operator.yaml
kubectl create -f $(pwd)/operator.yaml
if [[ "$?" -ne 0 ]] ; then echo "Failed creating operator deployment." ; exit $? ; fi

echo "Labelling CRDs"
kubectl get crds -o custom-columns=:metadata.name --no-headers | grep certmanager.k8s.io | while read crdName ; do kubectl label --overwrite crd $crdName app=ibm-cert-manager-controller ; done

echo "Creating CR"
kubectl create -f $(pwd)/crds/operator.ibm.com_v1alpha1_certmanager_cr.yaml
if [[ "$?" -ne 0 ]] ; then echo "Failed creating cr." ; exit $? ; fi

echo "Creating pull secret for certmanager deploy"
kubectl create secret docker-registry pull-secret-2 -n cert-manager --docker-server=hyc-cloud-private-edge-docker-local.artifactory.swg-devops.com --docker-username=$1 --docker-password=$2
if [[ "$?" -ne 0 ]] ; then echo "Failed creating image pull secret" ; exit $? ; fi
