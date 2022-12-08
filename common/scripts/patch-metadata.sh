#!/usr/bin/env bash
#
# Copyright 2022 IBM Corporation
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

yq=${1}
crdDir=${2:-config/crd/bases}

# add unapproved api annotation to v1alpha1 APIs which have domian k8s.io
# for f in "$crdDir"/certmanager.k8s.io_*; do
#     "${yq}" eval '.metadata.annotations."api-approved.kubernetes.io" = "unapproved"' "${f}" -i
# done

# add labels to resources
"${yq}" eval '.metadata.labels."app.kubernetes.io/instance" = "ibm-cert-manager-operator"' config/rbac/role.yaml -i
"${yq}" eval '.metadata.labels."app.kubernetes.io/managed-by" = "ibm-cert-manager-operator"' config/rbac/role.yaml -i
"${yq}" eval '.metadata.labels."app.kubernetes.io/name" = "cert-manager"' config/rbac/role.yaml -i
