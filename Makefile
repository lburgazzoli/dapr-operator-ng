
IMAGE_TAG_BASE ?= quay.io/lburgazzoli/dapr-operator-ng
IMG_VERSION ?= latest
IMG ?= ${IMAGE_TAG_BASE}:${IMG_VERSION}

LINT_GOGC := 10
LINT_DEADLINE := 10m

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
LOCALBIN ?= ${PROJECT_PATH}/bin

HELM_CHART_REPO ?= https://dapr.github.io/helm-charts
HELM_CHART ?= dapr
HELM_CHART_VERSION ?= 1.11.0
HELM_CHART_URL ?= https://raw.githubusercontent.com/dapr/helm-charts/master/dapr-$(HELM_CHART_VERSION).tgz

## Tool Versions
CODEGEN_VERSION := v0.27.4
KUSTOMIZE_VERSION ?= v5.0.1
CONTROLLER_TOOLS_VERSION ?= v0.12.1
KIND_VERSION ?= v0.20.0
LINTER_VERSION ?= v1.52.2
OPERATOR_SDK_VERSION ?= v1.31.0

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
LINTER ?= $(LOCALBIN)/golangci-lint
GOIMPORT ?= $(LOCALBIN)/goimports
YQ ?= $(LOCALBIN)/yq
KIND ?= $(LOCALBIN)/kind

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Dapr

.PHONY: update/dapr
update/dapr: ## Update the helm chart.
	$(PROJECT_PATH)/hack/scripts/update_helm_chart.sh $(PROJECT_PATH) $(HELM_CHART_URL)

##@ Development

.PHONY: manifests
manifests: codegen-tools-install ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(PROJECT_PATH)/hack/scripts/gen_crd.sh $(PROJECT_PATH)

.PHONY: generate
generate: codegen-tools-install ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(PROJECT_PATH)/hack/scripts/gen_res.sh $(PROJECT_PATH)
	$(PROJECT_PATH)/hack/scripts/gen_client.sh $(PROJECT_PATH)

.PHONY: fmt
fmt: goimport ## Run go fmt, gomiport against code.
	$(GOIMPORT) -l -w .
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet ## Run tests.
	go test -ldflags="$(GOLDFLAGS)" -v ./pkg/... ./internal/...

.PHONY: test/e2e
test/e2e: manifests generate fmt vet ## Run e2e tests.
	go test -ldflags="$(GOLDFLAGS)" -v ./test/e2e/...

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -ldflags="$(GOLDFLAGS)" -o bin/dapr-control-plane cmd/main.go

.PHONY: run
run: ## Run a controller from your host.
	go run -ldflags="$(GOLDFLAGS)" cmd/main.go run --leader-election=false --zap-devel

.PHONY: run/local
run/local: install ## Install and Run a controller from your host.
	go run -ldflags="$(GOLDFLAGS)" cmd/main.go run --leader-election=false --zap-devel


.PHONY: deps
deps:  ## Tidy up deps.
	go mod tidy


.PHONY: check/lint
check: check/lint

.PHONY: check/lint
check/lint: golangci-lint
	@$(LINTER) run \
		--config .golangci.yml \
		--out-format tab \
		--skip-dirs etc \
		--deadline $(LINT_DEADLINE) \
		--verbose

.PHONY: check/lint/fix
check/lint/fix: golangci-lint
	@$(LINTER) run \
		--config .golangci.yml \
		--out-format tab \
		--skip-dirs etc \
		--deadline $(LINT_DEADLINE) \
		--fix

.PHONY: docker/build
docker/build: test ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker/push
docker/push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

.PHONY: docker/push/kind
docker/push/kind: docker/build ## Load docker image in kind.
	kind load docker-image ${IMG}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/deploy/standalone | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/deploy/standalone | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -


.PHONY: deploy/e2e
deploy/e2e: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/deploy/e2e | kubectl apply -f -
.PHONY: deploy/kind
deploy/kind: manifests kustomize kind ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	kind load docker-image ${IMG}
	$(KUSTOMIZE) build config/deploy/standalone | kubectl apply -f -


.PHONY: build/bundle
build/bundle: kustomize operator-sdk yq
	rm -rf $(PROJECT_PATH)/bundles
	mkdir -p $(PROJECT_PATH)/bundles
	cd  $(PROJECT_PATH)/bundles

	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle \
	  --package dapr-operator-ng \
	  --version 0.0.1 \
	  --channels "alpha" \
	  --default-channel "alpha"

	$(YQ) -i \
		'.metadata.annotations.containerImage = .spec.install.spec.deployments[0].spec.template.spec.containers[0].image' \
		bundle/manifests/dapr-operator-ng.clusterserviceversion.yaml
	$(OPERATOR_SDK) bundle validate ./bundle

##@ Build Dependencies

## Location to install dependencies to
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

## Tool Versions
KUSTOMIZE_VERSION ?= v5.0.1
CONTROLLER_TOOLS_VERSION ?= v0.12.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)


.PHONY: golangci-lint
golangci-lint: $(LINTER)
$(LINTER): $(LOCALBIN)
	@test -s $(LOCALBIN)/golangci-lint || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(LINTER_VERSION)

.PHONY: goimport
goimport: $(GOIMPORT)
$(GOIMPORT): $(LOCALBIN)
	@test -s $(LOCALBIN)/goimport || \
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@latest

.PHONY: yq
yq: $(YQ)
$(YQ): $(LOCALBIN)
	@test -s $(LOCALBIN)/yq || \
	GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/v4@latest


.PHONY: kind
kind: $(KIND)
$(KIND): $(LOCALBIN)
	@test -s $(LOCALBIN)/kind || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kind@$(KIND_VERSION)


.PHONY: codegen-tools-install
codegen-tools-install: $(LOCALBIN)
	@echo "Installing code gen tools"
	$(PROJECT_PATH)/hack/scripts/install_gen_tools.sh $(PROJECT_PATH) $(CODEGEN_VERSION) $(CONTROLLER_TOOLS_VERSION)

.PHONY: operator-sdk
operator-sdk: $(LOCALBIN)
	@echo "Installing operator-sdk"
	$(PROJECT_PATH)/hack/scripts/install_operator_sdk.sh $(PROJECT_PATH) $(OPERATOR_SDK_VERSION)
