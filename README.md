# ibm-cert-manager-operator

Operator used to manage the ibm cert-manager service

## Supported platforms

### Platforms

- OCP 3.11
- OCP 4.1
- OCP 4.2
- OCP 4.3

### Operating Systems

- Linux amd64
- RHEL amd64
- RHEL ppc64le
- RHEL s390x

## Operator versions

- 3.5.0

## Prerequisites

1. Kubernetes 1.11 must be installed
1. OpenShift 3.11+ must be installed

## Documentation

For installation and configuration, see [IBM Knowledge Center link].

### Developer guide

Information about building and testing the operator.
- Dev quick start
  1. Follow the [ODLM guide](https://github.com/IBM/operand-deployment-lifecycle-manager/blob/master/docs/install/common-service-integration.md#end-to-end-test)

- Debugging the operator
  1. Check the CertManager CR

    ````
    kubectl get certmanager
    kubectl describe certmanager <certmanager CR name>
    ````

  1. Look at the logs of the cert-manager-operator pod for errors

    ````
    kubectl get po -n <namespace>
    kubectl logs -n <namespace> <cert-manager-operator pod name>
    ````

NOTE: cert-manager service (operand) is a singleton and no more than one instance of cert-manager-controller can be run within the same cluster.

## Licensing

Licensing
Copyright 2020 IBM Corp.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at [apache's website](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
