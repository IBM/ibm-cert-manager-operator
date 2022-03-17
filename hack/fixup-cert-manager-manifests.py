#!/usr/bin/env python3
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

"""
Manipulate the upstream cert-manager yaml manifests so that they are more
compatible with OLM.
* Reduces the combined size of CRDs to <1MiB (the size limit of a configmap)
  This is a work around for https://github.com/operator-framework/operator-lifecycle-manager/issues/1523
  Adapted from
  https://raw.githubusercontent.com/kubevirt/hyperconverged-cluster-operator/c6b425961feb0f350655ccfa7401336b30de66ab/hack/strip_old_descriptions.py
  in https://github.com/kubevirt/hyperconverged-cluster-operator/pull/1396
  Retain only the description fields of the stored API version with the exception of descriptions related to podTemplate,
  because those are so verbose and repeated multiple times.
Usage:
  hack/fixup-cert-manager-manifests  < build/cert-manager-1.4.0.upstream.yaml > build/cert-manager-1.4.0.olm.yaml
"""
import sys

import yaml

rubbish = ("description",)


def remove_descriptions(obj, keep=True, context=None):
    """
    Recursively remove any field called "description"
    """
    if context == "podTemplate":
        keep = False

    if isinstance(obj, dict):
        obj = {
            key: remove_descriptions(value, keep, context=key)
            for key, value in obj.items()
            if keep or key not in rubbish
        }
    elif isinstance(obj, list):
        obj = [
            remove_descriptions(item, keep, context=None)
            for i, item in enumerate(obj)
        ]
    return obj


def remove_descriptions_from_non_storage_versions_in_crd(crd):
    """
    Remove the description fields from the non-stored CRD versions.
    """
    crd_versions = crd["spec"]["versions"]
    for i, crd_version in enumerate(crd_versions):
        crd_versions[i] = remove_descriptions(crd_version, keep=crd_version.get("storage"))


def main():
    """
    Strip duplicate description fields from all supplied CRD files.
    """
    for doc in yaml.safe_load_all(sys.stdin):
        if doc.get("kind", "") == "CustomResourceDefinition":
            remove_descriptions_from_non_storage_versions_in_crd(doc)
        yaml.safe_dump(doc, sys.stdout)
        sys.stdout.write("---\n")


if __name__ == "__main__":
    main()