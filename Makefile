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
# Specify whether this repo is build locally or not, default values is '1';
# If set to 1, then you need to also set 'DOCKER_USERNAME' and 'DOCKER_PASSWORD'
# environment variables before build the repo.
BUILD_LOCALLY ?= 1

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overriding the IMG and REGISTRY environment variable.
IMG ?= ibm-cert-manager-operator
ifeq ($(BUILD_LOCALLY),0)
REGISTRY ?= "hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom"
else
REGISTRY ?= "quay.io/swatitestorg"
endif

IMAGE_REPO ?= $(REGISTRY)

# Set the registry and tag for the operand/operator images
OPERAND_REGISTRY ?= $(REGISTRY)
CERT_MANAGER_OPERAND_TAG ?= 0.10.7
CONFIGMAP_WATCHER_OPERAND_TAG ?= 3.3.4

# Current Operator version
VCS_URL ?= https://github.com/IBM/ibm-cert-manager-operator
VCS_REF ?= $(shell git rev-parse HEAD)
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

# Current Operator image name
OPERATOR_IMAGE_NAME ?= ibm-cert-manager-operator
# Current Operator bundle image name
BUNDLE_IMAGE_NAME ?= ibm-cert-manager-operator-bundle
# Current Operator version
OPERATOR_VERSION ?= $(VERSION)

# Default bundle image tag
# BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
# IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
    export CONFIG_DOCKER_TARGET_QUAY = config-docker-quay
endif

include common/Makefile.common.mk


##@ Development

check: lint-all ## Check all files lint error

code-dev: ## Run the default dev commands which are the go tidy, fmt, vet then execute the $ make code-gen
	@echo Running the common required commands for developments purposes
	- make generate-all
	- make code-tidy
	- make code-fmt
	- make code-vet
	@echo Running the common required commands for code delivery
	- make check
	- make test

manager: generate code-fmt code-vet ## Build manager binary
	go build -o bin/manager main.go

run: generate code-fmt code-vet manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	go run ./main.go -v=2

install: manifests kustomize ## Install CRDs into a cluster
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from a cluster
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMAGE_REPO)/$(OPERATOR_IMAGE_NAME):$(OPERATOR_VERSION)
	$(KUSTOMIZE) build config/default | kubectl apply -f -


##@ Generate code and manifests

manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=ibm-cert-manager-operator webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

bundle-manifests: ## Generate bundle manifests
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle \
	-q --overwrite --version $(OPERATOR_VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

generate-all: manifests ## Generate bundle manifests, metadata
	operator-sdk generate kustomize manifests -q
	- make bundle-manifests CHANNELS=beta,stable-v1 DEFAULT_CHANNEL=stable-v1

##@ Test

test: generate fmt vet manifests ## Run tests
	go test ./... -coverprofile cover.out


##@ Build

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
endif

build: ## Build operator binary
	@echo "Building the ibm-cert-manager-operator binary"
	@CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o bin/manager main.go

build-bundle-image: ## Build the operator bundle image.
	docker build -f bundle.Dockerfile -t $(REGISTRY)/$(BUNDLE_IMAGE_NAME)-$(LOCAL_ARCH):$(VERSION) .

build-image-amd64: ## Build amd64 operator image
	@docker build -t $(REGISTRY)/$(IMG)-amd64:$(VERSION) -f Dockerfile .

build-image-ppc64le: ## Build ppcle64 operator image
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(REGISTRY)/$(IMG)-ppc64le:$(VERSION) -f Dockerfile.ppc64le .

build-image-s390x: ## Build s390x operator image
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(REGISTRY)/$(IMG)-s390x:$(VERSION) -f Dockerfile.s390x .

push-image-amd64: $(CONFIG_DOCKER_TARGET) build-image-amd64 ## Push amd64 operator image
	@docker push $(REGISTRY)/$(IMG)-amd64:$(VERSION)

push-image-ppc64le: $(CONFIG_DOCKER_TARGET) build-image-ppc64le ## Push ppc64le operator image
	@docker push $(REGISTRY)/$(IMG)-ppc64le:$(VERSION)

push-image-s390x: $(CONFIG_DOCKER_TARGET) build-image-s390x ## Push s390x operator image
	@docker push $(REGISTRY)/$(IMG)-s390x:$(VERSION)

push-bundle-image: $(CONFIG_DOCKER_TARGET) build-bundle-image ## Push operator bundle image
	@docker push $(IMAGE_REPO)/$(BUNDLE_IMAGE_NAME)-$(ARCH):$(VERSION)


##@ Release

images: push-image-amd64 push-image-ppc64le push-image-s390x push-bundle-image multiarch-image multiarch-bundle-image ## Generate all images

multiarch-image: ## Generate multiarch images for operator image
	@curl -L -o /tmp/manifest-tool https://github.com/estesp/manifest-tool/releases/download/v1.0.0/manifest-tool-linux-amd64
	@chmod +x /tmp/manifest-tool
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(REGISTRY)/$(IMG)-ARCH:$(VERSION) --target $(REGISTRY)/$(IMG) --ignore-missing
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(REGISTRY)/$(IMG)-ARCH:$(VERSION) --target $(REGISTRY)/$(IMG):$(VERSION) --ignore-missing

multiarch-bundle-image: ## Generate multiarch images for operator bundle image
	@curl -L -o /tmp/manifest-tool https://github.com/estesp/manifest-tool/releases/download/v1.0.0/manifest-tool-linux-amd64
	@chmod +x /tmp/manifest-tool
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(REGISTRY)/$(BUNDLE_IMAGE_NAME):$(VERSION) --target $(REGISTRY)/$(BUNDLE_IMAGE_NAME) --ignore-missing
	#/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(REGISTRY)/$(BUNDLE_IMAGE_NAME)-ARCH:$(VERSION) --target $(REGISTRY)/$(BUNDLE_IMAGE_NAME):$(VERSION) --ignore-missing

##@ Help

help: ## Display this help
	@echo "Usage:\n  make \033[36m<target>\033[0m"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
