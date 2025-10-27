# Copyright 2019 The Kubernetes Authors.
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

############################################################
# GKE section
############################################################
PROJECT ?= oceanic-guard-191815
ZONE    ?= us-west1-a
CLUSTER ?= prow

activate-serviceaccount:
ifdef GOOGLE_APPLICATION_CREDENTIALS
	@gcloud auth activate-service-account --key-file="$(GOOGLE_APPLICATION_CREDENTIALS)"
endif

get-cluster-credentials: activate-serviceaccount
	@gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

config-docker:
	@echo "Configuring docker for building images"
	@if [ -z "$(DOCKER_USER)" ] || [ -z "$(DOCKER_PASS)" ]; then \
			echo "Error: DOCKER_USER and DOCKER_PASS must be defined"; \
			exit 1; \
		fi
	docker login -u $(DOCKER_USER) -p $(DOCKER_PASS) $(DOCKER_REGISTRY); \

############################################################
# lint section
############################################################

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

FINDFILES=find . \( -path ./.git -o -path ./.github \) -prune -o -type f
XARGS = xargs -0 ${XARGS_FLAGS}
CLEANXARGS = xargs ${XARGS_FLAGS}

GOIMPORTS_BIN ?= goimports
HADOLINT_BIN ?= hadolint
SHELLCHECK_BIN ?= shellcheck
YAMLLINT_BIN ?= $(shell command -v yamllint 2>/dev/null)
HELM_BIN ?= $(shell command -v helm 2>/dev/null)
MDL_BIN ?= $(shell command -v mdl 2>/dev/null)
AWESOME_BOT_BIN ?= $(shell command -v awesome_bot 2>/dev/null)
SASS_LINT_BIN ?= $(shell command -v sass-lint 2>/dev/null)
TSLINT_BIN ?= $(shell command -v tslint 2>/dev/null)
PROTOTOOL_BIN ?= $(shell command -v prototool 2>/dev/null)

lint-dockerfiles: | hadolint
	@if [ -z "$(HADOLINT_BIN)" ]; then \
		echo "Skipping lint-dockerfiles: hadolint not available."; \
	else \
		${FINDFILES} -name 'Dockerfile*' -print0 | ${XARGS} $(HADOLINT_BIN) --config ./common/config/.hadolint.yml --failure-threshold error; \
	fi

lint-scripts:
lint-scripts: | shellcheck
	@if [ -z "$(SHELLCHECK_BIN)" ]; then \
		echo "Skipping lint-scripts: shellcheck not available."; \
	else \
		${FINDFILES} -name '*.sh' -print0 | ${XARGS} $(SHELLCHECK_BIN); \
	fi

lint-yaml:
	@if [ -z "$(YAMLLINT_BIN)" ]; then \
		echo "Skipping lint-yaml: yamllint not available."; \
	else \
		${FINDFILES} \( -name '*.yml' -o -name '*.yaml' \) -print0 | { ${XARGS} grep -L -e "{{" || true; } | ${CLEANXARGS} $(YAMLLINT_BIN) -c ./common/config/.yamllint.yml; \
	fi

lint-helm:
	@if [ -z "$(HELM_BIN)" ]; then \
		echo "Skipping lint-helm: helm not available."; \
	else \
		${FINDFILES} -name 'Chart.yaml' -print0 | ${XARGS} -L 1 dirname | ${CLEANXARGS} $(HELM_BIN) lint; \
	fi

lint-copyright-banner:
	@${FINDFILES} \( -name '*.go' -o -name '*.cc' -o -name '*.h' -o -name '*.proto' -o -name '*.py' -o -name '*.sh' \) \( ! \( -name '*.gen.go' -o -name '*.pb.go' -o -name '*_pb2.py' \) \) -print0 |\
		${XARGS} common/scripts/lint_copyright_banner.sh

lint-go:
	# @${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' \) \) -print0 | ${XARGS} common/scripts/lint_go.sh

lint-python:
	@${FINDFILES} -name '*.py' \( ! \( -name '*_pb2.py' \) \) -print0 | ${XARGS} autopep8 --max-line-length 160 --exit-code -d

lint-markdown:
	@if [ -z "$(MDL_BIN)" ]; then \
		echo "Skipping lint-markdown: mdl not available."; \
	else \
		${FINDFILES} -name '*.md' -print0 | ${XARGS} $(MDL_BIN) --ignore-front-matter --style common/config/mdl.rb; \
	fi
ifdef MARKDOWN_LINT_WHITELIST
	@if [ -z "$(AWESOME_BOT_BIN)" ]; then \
		echo "Skipping awesome_bot markdown lint: awesome_bot not available."; \
	else \
		${FINDFILES} -name '*.md' -print0 | ${XARGS} $(AWESOME_BOT_BIN) --skip-save-results --allow_ssl --allow-timeout --allow-dupe --allow-redirect --white-list ${MARKDOWN_LINT_WHITELIST}; \
	fi
else
	@if [ -z "$(AWESOME_BOT_BIN)" ]; then \
		echo "Skipping awesome_bot markdown lint: awesome_bot not available."; \
	else \
		${FINDFILES} -name '*.md' -print0 | ${XARGS} $(AWESOME_BOT_BIN) --skip-save-results --allow_ssl --allow-timeout --allow-dupe --allow-redirect; \
	fi
endif

lint-sass:
	@if [ -z "$(SASS_LINT_BIN)" ]; then \
		echo "Skipping lint-sass: sass-lint not available."; \
	else \
		${FINDFILES} -name '*.scss' -print0 | ${XARGS} $(SASS_LINT_BIN) -c common/config/sass-lint.yml --verbose; \
	fi

lint-typescript:
	@if [ -z "$(TSLINT_BIN)" ]; then \
		echo "Skipping lint-typescript: tslint not available."; \
	else \
		${FINDFILES} -name '*.ts' -print0 | ${XARGS} $(TSLINT_BIN) -c common/config/tslint.json; \
	fi

lint-protos:
	@if [ -z "$(PROTOTOOL_BIN)" ]; then \
		echo "Skipping lint-protos: prototool not available."; \
	else \
		$(FINDFILES) -name '*.proto' -print0 | $(XARGS) -L 1 $(PROTOTOOL_BIN) lint --protoc-bin-path=/usr/bin/protoc; \
	fi

lint-all: lint-dockerfiles lint-scripts lint-yaml lint-helm lint-copyright-banner lint-go lint-python lint-markdown lint-sass lint-typescript lint-protos

format-go: | goimports
	@${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' \) \) -print0 | ${XARGS} $(GOIMPORTS_BIN) -w -local "github.com/IBM"

format-python:
	@${FINDFILES} -name '*.py' -print0 | ${XARGS} autopep8 --max-line-length 160 --aggressive --aggressive -i

format-protos:
	@$(FINDFILES) -name '*.proto' -print0 | $(XARGS) -L 1 prototool format -w

.PHONY: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-python lint-helm lint-markdown lint-sass lint-typescript lint-protos lint-all format-go format-python format-protos config-docker
