#!/usr/bin/env bash
#
# Copyright 2021 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

# using yq v4.3.1; ideally all the files can be specified in one command
# but the "-i" flag does not update all the files
yq -i eval-all '.metadata.annotations."api-approved.kubernetes.io" = "unapproved"' deploy/crds/certmanager.k8s.io_certificaterequests_crd.yaml
yq -i eval '.metadata.annotations."api-approved.kubernetes.io" = "unapproved"' deploy/crds/certmanager.k8s.io_certificates_crd.yaml
yq -i eval '.metadata.annotations."api-approved.kubernetes.io" = "unapproved"' deploy/crds/certmanager.k8s.io_challenges_crd.yaml
yq -i eval '.metadata.annotations."api-approved.kubernetes.io" = "unapproved"' deploy/crds/certmanager.k8s.io_clusterissuers_crd.yaml
yq -i eval '.metadata.annotations."api-approved.kubernetes.io" = "unapproved"' deploy/crds/certmanager.k8s.io_issuers_crd.yaml
yq -i eval '.metadata.annotations."api-approved.kubernetes.io" = "unapproved"' deploy/crds/certmanager.k8s.io_orders_crd.yaml
yq -i eval '.metadata.annotations."api-approved.kubernetes.io" = "unapproved"' deploy/crds/operator.ibm.com_certmanagers_crd.yaml

labels_file="common/scripts/labels.yaml"

yq -i eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' \
    deploy/crds/certmanager.k8s.io_certificaterequests_crd.yaml "$labels_file"

yq -i eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' \
    deploy/crds/certmanager.k8s.io_certificates_crd.yaml "$labels_file"

yq -i eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' \
    deploy/crds/certmanager.k8s.io_challenges_crd.yaml "$labels_file"

yq -i eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' \
    deploy/crds/certmanager.k8s.io_clusterissuers_crd.yaml "$labels_file"

yq -i eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' \
    deploy/crds/certmanager.k8s.io_issuers_crd.yaml "$labels_file"

yq -i eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' \
    deploy/crds/certmanager.k8s.io_orders_crd.yaml "$labels_file"

yq -i eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' \
    deploy/crds/operator.ibm.com_certmanagers_crd.yaml "$labels_file"
