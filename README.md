# ibm-cert-manager-operator

Operator used to manage the IBM certificate manager service.

## Supported platforms

- OCP 3.11
- OCP 4.1
- OCP 4.2
- OCP 4.3

## Operating Systems

- Linux amd64
- RHEL amd64
- RHEL ppc64le
- RHEL s390x

## Operator versions

- 3.5.0

## Prerequisites

1. Kubernetes 1.11 must be installed.
1. OpenShift 3.11+ must be installed.

## Documentation

For installation and configuration, see the [IBM Cloud Platform Common Services documentation](http://ibm.biz/cpcsdocs).

### Developer guide

Information about building and testing the operator.
- Developer quick start
  1. Follow the [ODLM guide](https://github.com/IBM/operand-deployment-lifecycle-manager/blob/master/docs/install/common-service-integration.md#end-to-end-test).

- Debugging the operator
  1. Check the certificate manager CR.

    ````
    kubectl get certmanager
    kubectl describe certmanager <certmanager CR name>
    ````

  1. Look at the logs of the cert-manager-operator pod for errors

    ````
    kubectl get po -n <namespace>
    kubectl logs -n <namespace> <cert-manager-operator pod name>
    ````

**NOTE:** The certificate manager service (operand) is a singleton. Do not run more than one instance of `cert-manager-controller` within the same cluster.

# SecurityContextConstraints Requirements

The cert-manager service supports running under the OpenShift Container Platform default restricted security context constraints.
