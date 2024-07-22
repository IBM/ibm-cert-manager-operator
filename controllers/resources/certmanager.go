//
// Copyright 2022 IBM Corporation
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

package resources

// base on doc https://www.ibm.com/docs/en/cpfs?topic=services-configuring-foundational-by-using-custom-resource#cert_resources
const CertManagerConfigCR = `
apiVersion: operator.ibm.com/v1
kind: CertManagerConfig
metadata:
  labels:
    app.kubernetes.io/instance: ibm-cert-manager-operator
    app.kubernetes.io/managed-by: ibm-cert-manager-operator
    app.kubernetes.io/name: cert-manager
  name: default
spec:
  certManagerCAInjector:
    resources:
      limits:
        cpu: 100m
        memory: 1000Mi
      requests:
        cpu: 30m
        memory: 500Mi
        ephemeral-storage: 256Mi
  certManagerController:
    resources:
      limits:
        cpu: 80m
        memory: 1010Mi
      requests:
        cpu: 20m
        memory: 230Mi
        ephemeral-storage: 510Mi
  certManagerWebhook:
    resources:
      limits:
        cpu: 60m
        memory: 100Mi
      requests:
        cpu: 30m
        memory: 40Mi
        ephemeral-storage: 256Mi
  disableHostNetwork: true
  enableCertRefresh: true
  enableWebhook: true
  imageRegistry: icr.io/cpopen/cpfs
  license:
    accept: false
  version: 4.2.7
status:
  certManagerConfigStatus: ''
`
