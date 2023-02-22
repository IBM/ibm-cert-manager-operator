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

const CertManagerCR = `
apiVersion: operator.ibm.com/v1
kind: CertManager
metadata:
  labels:
    app.kubernetes.io/instance: ibm-cert-manager-operator
    app.kubernetes.io/managed-by: ibm-cert-manager-operator
    app.kubernetes.io/name: cert-manager
  name: default
spec:
  disableHostNetwork: true
  enableCertRefresh: true
  enableWebhook: true
  imageRegistry: icr.io/cpopen/cpfs
  version: 4.0.0
status:
  certManagerStatus: ''
`
