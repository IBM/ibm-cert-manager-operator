#!/bin/bash
#
# Copyright 2020 IBM Corporation
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
# Create a new version of the operator and update the required files with the new version number.
#
# Run this script from the parent dir by typing "scripts/bump-up-csv.sh"
#

SED="sed"
unamestr=$(uname)
if [[ "$unamestr" == "Darwin" ]] ; then
    SED=gsed
    type $SED >/dev/null 2>&1 || {
        echo >&2 "$SED it's not installed. Try: brew install gnu-sed" ;
        exit 1;
    }
fi

# check the input parms
if [ -z "${1}" ]; then
    echo "Usage:   $0 [new CSV version]"
    exit 1
fi

if [ ! -f "$(command -v yq 2> /dev/null)" ]; then
    echo "[ERROR] yq command not found"
    exit 1
fi

OPERATOR_NAME=ibm-cert-manager-operator
NEW_CSV_VERSION=${1}

DEPLOY_DIR=${DEPLOY_DIR:-deploy}
BUNDLE_DIR=${BUNDLE_DIR:-deploy/olm-catalog/${OPERATOR_NAME}}
# get the version number for the current/last CSV
LAST_CSV_DIR=$(find "${BUNDLE_DIR}" -maxdepth 1 -type d | sort | tail -1)
LAST_CSV_VERSION=$(basename "${LAST_CSV_DIR}")
NEW_CSV_DIR=${LAST_CSV_DIR//${LAST_CSV_VERSION}/${NEW_CSV_VERSION}}

PREVIOUS_CSV_DIR=$(find "${BUNDLE_DIR}" -maxdepth 1 -type d | sort | tail -2 | head -1)
PREVIOUS_CSV_VERSION=$(basename "${PREVIOUS_CSV_DIR}")

if [ "${LAST_CSV_VERSION}" == "${NEW_CSV_VERSION}" ]; then
    echo "Last CSV version is already at ${NEW_CSV_VERSION}"
    exit 1
fi

echo "******************************************"
echo " PREVIOUS_CSV_VERSION:  $PREVIOUS_CSV_VERSION"
echo " CURRENT_CSV_VERSION:   $LAST_CSV_VERSION"
echo " NEW_CSV_VERSION:       $NEW_CSV_VERSION"
echo "******************************************"
echo "Does the above look correct? (y/n) "
read -r ANSWER
if [[ "$ANSWER" != "y" ]]
then
  echo "Not going to bump up CSV version"
  exit 1
fi

echo -e "\n[INFO] Bumping up CSV version from ${LAST_CSV_VERSION} to ${NEW_CSV_VERSION}\n"

cp -rfv "${LAST_CSV_DIR}" "${NEW_CSV_DIR}"
OLD_CSV_FILE=$(find "${NEW_CSV_DIR}" -type f -name '*.clusterserviceversion.yaml' | head -1)
NEW_CSV_FILE=${OLD_CSV_FILE//${LAST_CSV_VERSION}.clusterserviceversion.yaml/${NEW_CSV_VERSION}.clusterserviceversion.yaml}
if [ -f "${OLD_CSV_FILE}" ]; then
    mv -v "${OLD_CSV_FILE}" "${NEW_CSV_FILE}"
fi

echo -e "\n[INFO] Updating ${NEW_CSV_FILE} from ${OLD_CSV_FILE}"

REPLACES_VERSION=$(yq r "${NEW_CSV_FILE}" "metadata.name")
CURR_TIME=$(TZ=":UTC" date  +'%FT%R:00Z')

#---------------------------------------------------------
# update the CSV file
#---------------------------------------------------------
$SED -e "s|name: ${OPERATOR_NAME}\(.*\)${LAST_CSV_VERSION}|name: ${OPERATOR_NAME}\1${NEW_CSV_VERSION}|" -i "${NEW_CSV_FILE}"
$SED -e "s|olm.skipRange: \(.*\)${LAST_CSV_VERSION}\(.*\)|olm.skipRange: \1${NEW_CSV_VERSION}\2|" -i "${NEW_CSV_FILE}"
$SED -e "s|image: \(.*\)${OPERATOR_NAME}\(.*\)|image: \1${OPERATOR_NAME}:latest|" -i "${NEW_CSV_FILE}"
$SED -e "s|replaces: ${OPERATOR_NAME}\(.*\)${PREVIOUS_CSV_VERSION}|replaces: ${REPLACES_VERSION}|" -i "${NEW_CSV_FILE}"
# update the operator version - version: 3.6.2
$SED -e "s|version: ${LAST_CSV_VERSION}|version: ${NEW_CSV_VERSION}|" -i "${NEW_CSV_FILE}"
# update the 'version' field for each CR in the CSV - "version": "3.6.3"
$SED -e "s|\"version\": \"${LAST_CSV_VERSION}\"|\"version\": \"${NEW_CSV_VERSION}\"|g" -i "${NEW_CSV_FILE}"
# update the 'createdAt' date. example:
#   createdAt: "2020-06-29T21:46:35Z"
#                           YYYY    -    MM    -    DD    T    HH    :   MM     :   SS     Z
#                       +----------+ +--------+ +--------+ +--------+ +--------+ +--------+
$SED -e "s|createdAt: \"\([0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}T[0-9]\{2\}:[0-9]\{2\}:[0-9]\{2\}Z\)\"|createdAt: \"${CURR_TIME}\"|" -i "${NEW_CSV_FILE}"

#---------------------------------------------------------
# update package.yaml
#---------------------------------------------------------
PACKAGE_YAML=${BUNDLE_DIR}/${OPERATOR_NAME}.package.yaml
if ! [ -f "${PACKAGE_YAML}" ]; then
    echo "[WARN] ${PACKAGE_YAML} does not exist."
    exit 1
fi
echo -e "\n[INFO] Updating 'dev' channel in ${PACKAGE_YAML}"
NEW_VERSION=$(yq r "${NEW_CSV_FILE}" "metadata.name")
yq w -i "${PACKAGE_YAML}" "channels.(name==dev).currentCSV" "${NEW_VERSION}" 

# remove the leading spaces added by "yq"
$SED -e "s|  - currentCSV:|- currentCSV:|g" -i "${PACKAGE_YAML}"
$SED -e "s|    name:|  name:|g" -i "${PACKAGE_YAML}"

#---------------------------------------------------------
# update operator.yaml
#---------------------------------------------------------
OPERATOR_YAML=${DEPLOY_DIR}/operator.yaml
if ! [ -f "${OPERATOR_YAML}" ]; then
    echo "[WARN] ${OPERATOR_YAML} does not exist."
    exit 1
fi
echo -e "\n[INFO] Updating 'image tag' in ${OPERATOR_YAML}"
$SED -e "s|image: \(.*\)${OPERATOR_NAME}\(.*\)|image: \1${OPERATOR_NAME}:latest|" -i "${OPERATOR_YAML}"

#---------------------------------------------------------
# update the version in certmanager CR yaml file
#---------------------------------------------------------
echo -e "\n[INFO] Updating 'version' in certmanager CR yaml"
$SED -e "s|version: \"${LAST_CSV_VERSION}\"|version: \"${NEW_CSV_VERSION}\"|" -i "${DEPLOY_DIR}/crds/operator.ibm.com_v1alpha1_certmanager_cr.yaml"

#---------------------------------------------------------
# update version.go
#---------------------------------------------------------
VERSION_GO="version/version.go"
if ! [ -f "${VERSION_GO}" ]; then
    echo "[WARN] ${VERSION_GO} does not exist."
    exit 1
fi
echo -e "\n[INFO] Updating 'version' in ${VERSION_GO}"
$SED -e "s|Version\(.*\)${LAST_CSV_VERSION}\(.*\)|Version\1${NEW_CSV_VERSION}\2|" -i "${VERSION_GO}"

#---------------------------------------------------------
# update .osdk-scorecard.yaml
#---------------------------------------------------------
if [ -f ".osdk-scorecard.yaml" ]; then
    echo -e "\n[INFO] Updating 'version' in .osdk-scorecard.yaml"
    $SED -e "s|${LAST_CSV_VERSION}|${NEW_CSV_VERSION}|g" -i .osdk-scorecard.yaml
fi
