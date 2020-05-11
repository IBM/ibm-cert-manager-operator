# ibm-cert-manager-operator

> **Important:** Do not install this operator directly. Only install this operator using the IBM Common Services Operator. For more information about installing this operator and other Common Services operators, see [Installer documentation](http://ibm.biz/cpcs_opinstall). If you are using this operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak to learn more about how to install and use the operator service. For more information about IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).

You can use the ibm-cert-manager-operator to install the IBM certificate manager service for the IBM Cloud Platform Common Services. You can use IBM certificate manager service to issue and manage x509 certificates from various sources, such as Let’s Encrypt and Hashicorp Vault, a simple signing key pair, or self-signed. It ensures certificates are valid and up to date and will renew certificates before they expire.

For more information about the available IBM Cloud Platform Common Services, see the [IBM Knowledge Center](http://ibm.biz/cpcsdocs).

## Supported platforms

Red Hat OpenShift Container Platform 4.2 or newer installed on one of the following platforms:

- Linux x86_64
- Linux on Power (ppc64le)
- Linux on IBM Z and LinuxONE

## Operator versions

- 3.5.0
- 3.6.0

## Prerequisites

Before you install this operator, you need to first install the operator dependencies and prerequisites:

- For the list of operator dependencies, see the IBM Knowledge Center [Common Services dependencies documentation](http://ibm.biz/cpcs_opdependencies).

- For the list of prerequisites for installing the operator, see the IBM Knowledge Center [Preparing to install services documentation](http://ibm.biz/cpcs_opinstprereq).

## Documentation

To install the operator with the IBM Common Services Operator follow the the installation and configuration instructions within the IBM Knowledge Center.

- If you are using the operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak. For a list of IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).
- If you are using the operator with an IBM Containerized Software, see the IBM Cloud Platform Common Services Knowledge Center [Installer documentation](http://ibm.biz/cpcs_opinstall).

## SecurityContextConstraints Requirements

The IBM certificate manager service supports running with the OpenShift Container Platform 4.3 default restricted Security Context Constraints (SCCs).

For more information about the OpenShift Container Platform Security Context Constraints, see [Managing Security Context Constraints](https://docs.openshift.com/container-platform/4.3/authentication/managing-security-context-constraints.html).

OCP 4.3 restricted SCC:

```yaml
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities: null
apiVersion: security.openshift.io/v1
defaultAddCapabilities: null
fsGroup:
  type: MustRunAs
groups:
- system:authenticated
kind: SecurityContextConstraints
metadata:
  annotations:
    kubernetes.io/description: restricted denies access to all host features and requires
      pods to be run with a UID, and SELinux context that are allocated to the namespace.  This
      is the most restrictive SCC and it is used by default for authenticated users.
  creationTimestamp: "2020-03-27T15:01:00Z"
  generation: 1
  name: restricted
  resourceVersion: "6365"
  selfLink: /apis/security.openshift.io/v1/securitycontextconstraints/restricted
  uid: 6a77775c-a6d8-4341-b04c-bd826a67f67e
priority: null
readOnlyRootFilesystem: false
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: MustRunAsRange
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
```

### Developer guide

If, as a developer, you are looking to build and test this operator to try out and learn more about the operator and its capabilities, you can use the following developer guide. This guide provides commands for a quick install and initial validation for running the operator.

> **Important:** The following developer guide is provided as-is and only for trial and education purposes. IBM and IBM Support does not provide any support for the usage of the operator with this developer guide. For the official supported install and usage guide for the operator, see the the IBM Knowledge Center documentation for your IBM Cloud Pak or for IBM Cloud Platform Common Services.

### Quick start guide

- Follow the [IBM Common Services Operator guide](https://github.com/IBM/ibm-common-service-operator/blob/master/docs/install.md).

- Use the following quick start commands for building and testing the operator:

```bash
oc login -u <CLUSTER_USER> -p <CLUSTER_PASS> <CLUSTER_IP>:6443

export NAMESPACE=ibm-common-services
export BASE_DIR=deploy/olm-catalog/ibm-cert-manager-operator
export CSV_VERSION=<CSV_VERSION_TO_TEST>

make install
```

- To uninstall the operator installed using `make install`, run `make uninstall`.

### Debugging guide

Use the following commands to debug the operator:

- Check the certificate manager CR.

```bash
kubectl get certmanager
kubectl describe certmanager <certmanager CR name>
```

- Look at the logs of the cert-manager-operator pod for errors

```bash
kubectl get po -n <namespace>
kubectl logs -n <namespace> <cert-manager-operator pod name>
```

**NOTE:** The certificate manager service (operand) is a singleton. Do not run more than one instance of `cert-manager-controller` within the same cluster.

### End-to-End testing

For more instructions on how to run end-to-end testing with the Operand Deployment Lifecycle Manager, see [IBM Common Services Operator guide](https://github.com/IBM/ibm-common-service-operator/blob/master/docs/install.md).
