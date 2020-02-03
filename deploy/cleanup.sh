#!/bin/bash
kubectl delete deploy ibm-cert-manager-operator -n cert-manager
kubectl delete clusterrole ibm-cert-manager-operator
kubectl delete clusterolebinding ibm-cert-manager-operator
kubectl delete serviceaccount ibm-cert-manager-operator
kubectl delete secret pull-secret-1 -n cert-manager
kubectl delete secret pull-secret-2 -n cert-manager
kubectl delete crds certmanagers.operator.ibm.com
