# Copyright 2020 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This repo is build locally for dev/test by default;
# Override this variable in CI env.
BUILD_LOCALLY ?= 1

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overriding the IMG and REGISTRY environment variable.
IMG ?= ibm-cert-manager-operator
REGISTRY ?= "hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom"

# Set the registry and tag for the operand/operator images
OPERAND_REGISTRY ?= $(REGISTRY)
CERT_MANAGER_OPERAND_TAG ?= 0.12.0
CONFIGMAP_WATCHER_OPERAND_TAG ?= 3.4.0

# Github host to use for checking the source tree;
# Override this variable ue with your own value if you're working on forked repo.
GIT_HOST ?= github.com/IBM

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))

# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT := $(GOPATH)/bin
export GOBIN ?= $(GOBIN_DEFAULT)
TESTARGS_DEFAULT := "-v"
export TESTARGS ?= $(TESTARGS_DEFAULT)
DEST := $(GOPATH)/src/$(GIT_HOST)/$(BASE_DIR)
VERSION ?= $(shell cat ./version/version.go | grep "Version =" | awk '{ print $$3}' | tr -d '"')
CSV_VERSION ?= $(VERSION)

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
    TARGET_OS ?= linux
    XARGS_FLAGS="-r"
else ifeq ($(LOCAL_OS),Darwin)
    TARGET_OS ?= darwin
    XARGS_FLAGS=
else
    $(error "This system's OS $(LOCAL_OS) isn't recognized/supported")
endif

ARCH := $(shell uname -m)
LOCAL_ARCH := "amd64"
ifeq ($(ARCH),x86_64)
    LOCAL_ARCH="amd64"
else ifeq ($(ARCH),ppc64le)
    LOCAL_ARCH="ppc64le"
else ifeq ($(ARCH),s390x)
    LOCAL_ARCH="s390x"
else
    $(error "This system's ARCH $(ARCH) isn't recognized/supported")
endif

all: fmt check test coverage build images

include common/Makefile.common.mk

############################################################
# work section
############################################################
$(GOBIN):
	@echo "create gobin"
	@mkdir -p $(GOBIN)

work: $(GOBIN)

############################################################
# format section
############################################################

# All available format: format-go format-protos format-python
# Default value will run all formats, override these make target with your requirements:
#    eg: fmt: format-go format-protos
fmt: format-go format-protos format-python

############################################################
# check section
############################################################

check: lint

# All available linters: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-python lint-helm lint-markdown lint-sass lint-typescript lint-protos
# Default value will run all linters, override these make target with your requirements:
#    eg: lint: lint-go lint-yaml
lint: lint-all

############################################################
# test section
############################################################

test:
	@go test ${TESTARGS} ./...

############################################################
# coverage section
############################################################

coverage:
	@common/scripts/codecov.sh ${BUILD_LOCALLY}

############################################################
# install operator sdk section
############################################################

install-operator-sdk: 
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then ./common/scripts/install-operator-sdk.sh; fi

############################################################
# build section
############################################################

build: build-amd64 build-ppc64le build-s390x

build-amd64:
	@echo "Building the ${IMG} amd64 binary..."
	@GOARCH=amd64 common/scripts/gobuild.sh build/_output/bin/$(IMG) ./cmd/manager

build-ppc64le:
	@echo "Building the ${IMG} ppc64le binary..."
	@GOARCH=ppc64le common/scripts/gobuild.sh build/_output/bin/$(IMG)-ppc64le ./cmd/manager

build-s390x:
	@echo "Building the ${IMG} s390x binary..."
	@GOARCH=s390x common/scripts/gobuild.sh build/_output/bin/$(IMG)-s390x ./cmd/manager

local:
	@GOOS=darwin common/scripts/gobuild.sh build/_output/bin/$(IMG) ./cmd/manager

############################################################
# images section
############################################################

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
config-docker:
endif


build-image-amd64: build-amd64
	@docker build -t $(REGISTRY)/$(IMG)-amd64:$(VERSION) -f build/Dockerfile .

build-image-ppc64le: build-ppc64le
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(REGISTRY)/$(IMG)-ppc64le:$(VERSION) -f build/Dockerfile.ppc64le .

build-image-s390x: build-s390x
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(REGISTRY)/$(IMG)-s390x:$(VERSION) -f build/Dockerfile.s390x .

push-image-amd64: $(CONFIG_DOCKER_TARGET) build-image-amd64
	@docker push $(REGISTRY)/$(IMG)-amd64:$(VERSION)

push-image-ppc64le: $(CONFIG_DOCKER_TARGET) build-image-ppc64le
	@docker push $(REGISTRY)/$(IMG)-ppc64le:$(VERSION)

push-image-s390x: $(CONFIG_DOCKER_TARGET) build-image-s390x
	@docker push $(REGISTRY)/$(IMG)-s390x:$(VERSION)

############################################################
# multiarch-image section
############################################################

images: push-image-amd64 push-image-ppc64le push-image-s390x multiarch-image

multiarch-image:
	@curl -L -o /tmp/manifest-tool https://github.com/estesp/manifest-tool/releases/download/v1.0.3/manifest-tool-linux-amd64
	@chmod +x /tmp/manifest-tool
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(REGISTRY)/$(IMG)-ARCH:$(VERSION) --target $(REGISTRY)/$(IMG) --ignore-missing
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(REGISTRY)/$(IMG)-ARCH:$(VERSION) --target $(REGISTRY)/$(IMG):$(VERSION) --ignore-missing

############################################################
# CSV section
############################################################
csv:
	common/scripts/push_csv.sh

############################################################
# clean section
############################################################
clean:
	rm -f build/_output

############################################################
# Bump up CSV version
############################################################

.PHONY: bump-csv
bump-csv:
	@common/scripts/bump_up_csv.sh ${NEW_CSV_VERSION}

############################################################
# application section
############################################################

install: ## Install all resources (CR/CRD's, RBAC and Operator)
	@echo ....... Set environment variables ......
	# - export NAMESPACE=ibm-common-services
	# - export BASE_DIR=deploy/olm-catalog/ibm-cert-manager-operator
	# @echo ....... Creating namespace .......
	# - kubectl create namespace ${NAMESPACE}
	@echo ....... Applying CRDS and Operator .......
	- for crd in $(shell ls deploy/crds/*crd.yaml); do kubectl apply -f $${crd}; done
	@echo ....... Applying RBAC .......
	- kubectl apply -f deploy/service_account.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role_binding.yaml -n ${NAMESPACE}
	@echo ....... Applying Operator .......
	# - kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/olm-catalog/${BASE_DIR}/${CSV_VERSION}/${BASE_DIR}.v${CSV_VERSION}.clusterserviceversion.yaml -n ${NAMESPACE}
	@echo ....... Creating the Instance .......
	- for cr in $(shell ls deploy/crds/*_cr.yaml); do kubectl -n ${NAMESPACE} apply -f $${cr}; done

run_local:
	# @echo ....... Applying CRDS and Operator .......
	# - for crd in $(shell ls deploy/crds/*crd.yaml); do kubectl apply -f $${crd}; done
	@echo ....... Run operator locally ........
	- operator-sdk run local --watch-namespace=${NAMESPACE}

install_cr_local:
	@echo ....... Creating the Instance .......
	- for cr in $(shell ls deploy/crds/*_cr.yaml); do kubectl -n ${NAMESPACE} apply -f $${cr}; done

uninstall_cr_local:
	@echo ....... Creating the Instance .......
	- kubectl -n ${NAMESPACE} delete -f deploy/crds/operator.ibm.com_v1alpha1_certmanager_cr.yaml

uninstall: ## Uninstall all that all performed in the $ make install
	@echo ....... Uninstalling .......
	@echo ....... Deleting CR .......
	- for cr in $(shell ls deploy/crds/*_cr.yaml); do kubectl -n ${NAMESPACE} delete -f $${cr}; done
	@echo ....... Deleting Operator .......
	#- kubectl delete -f deploy/operator.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/olm-catalog/${BASE_DIR}/${CSV_VERSION}/${BASE_DIR}.v${CSV_VERSION}.clusterserviceversion.yaml -n ${NAMESPACE}
	@echo ....... Deleting CRDs.......
	- for crd in $(shell ls deploy/crds/*crd.yaml); do kubectl delete -f $${crd}; done
	@echo ....... Deleting Rules and Service Account .......
	- kubectl delete -f deploy/role_binding.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/role.yaml -n ${NAMESPACE}
	# @echo ....... Deleting namespace .......
	# - kubectl delete namespace ${NAMESPACE}

############################################################
# local dev section
############################################################

push-image-dev:
	make build-image-amd64 VERSION=dev
	docker push $(REGISTRY)/$(IMG)-amd64:dev

dev-csv:
	@common/scripts/replace-dev-images.sh \
		$(REGISTRY) \
		$(IMG) \
		deploy/olm-catalog/ibm-cert-manager-operator/${VERSION}/ibm-cert-manager-operator.v${VERSION}.clusterserviceversion.yaml

# deploys CSV using currently installed OLM on the cluster
# change pullPolicy value in pkg/resources/constants.go to always pull operands
run-csv:
	operator-sdk run packagemanifests \
		--operator-version ${VERSION} \
		--operator-namespace ibm-common-services \
		--olm-namespace openshift-operator-lifecycle-manager
		
	oc apply -f deploy/crds/operator.ibm.com_v1alpha1_certmanager_cr.yaml

cleanup-csv:
	oc delete -f deploy/crds/operator.ibm.com_v1alpha1_certmanager_cr.yaml

	operator-sdk cleanup packagemanifests \
		--operator-version ${VERSION} \
		--operator-namespace ibm-common-services \
		--olm-namespace openshift-operator-lifecycle-manager

.PHONY: all work build check lint test coverage images multiarch-image
