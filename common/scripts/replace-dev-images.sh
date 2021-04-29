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

REGISTRY=$1
IMG=$2
CSV_FILE=$3


# replace operator image
yq -i eval ".spec.install.spec.deployments[0].spec.template.spec.containers[0].image = \"${REGISTRY}/${IMG}-amd64:dev\"" \
	${CSV_FILE}

# replace operand images
yq -i eval ".spec.install.spec.deployments[0].spec.template.spec.containers[0].env[0].value = \"${REGISTRY}/icp-cert-manager-controller:dev\"" \
	${CSV_FILE}
yq -i eval ".spec.install.spec.deployments[0].spec.template.spec.containers[0].env[1].value = \"${REGISTRY}/icp-cert-manager-webhook:dev\"" \
	${CSV_FILE}
yq -i eval ".spec.install.spec.deployments[0].spec.template.spec.containers[0].env[2].value = \"${REGISTRY}/icp-cert-manager-cainjector:dev\"" \
	${CSV_FILE}
yq -i eval ".spec.install.spec.deployments[0].spec.template.spec.containers[0].env[3].value = \"${REGISTRY}/icp-cert-manager-acmesolver:dev\"" \
	${CSV_FILE}
