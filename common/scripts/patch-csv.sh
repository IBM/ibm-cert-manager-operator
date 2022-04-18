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

sed="sed"
unamestr=$(uname)
if [[ "$unamestr" == "Darwin" ]] ; then
    SED=gsed
    type $SED >/dev/null 2>&1 || {
        echo >&2 "$SED it's not installed. Try: brew install gnu-sed" ;
        exit 1;
    }
fi

prev_version=${1}
curr_version=${2}
csv=${3:-config/manifests/bases/ibm-cert-manager-operator.clusterserviceversion.yaml}

# add labels to resources
"${sed}" -e "s|replaces: ibm-cert-manager-operator\(.*\)|replaces: ibm-cert-manager-operator.${prev_version}|" -i "${csv}"
"${sed}" -e "s|olm.skipRange: <\(.*\)|olm.skipRange: <${curr_version}|" -i "${csv}"
