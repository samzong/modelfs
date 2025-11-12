.PHONY: help generate manifests run tidy controller-gen build docker-build docker-push helm-package helm-install helm-uninstall helm-lint

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
KIND_CLUSTER_NAME ?= modelfs
KIND_CONFIG ?= hack/kind-config.yaml
DATASET_NAMESPACE ?= dataset-system
DATASET_REPO ?= https://github.com/BaizeAI/dataset.git
DATASET_CLONE_DIR ?= third_party/dataset
DATASET_RELEASE ?= dataset
DATASET_HELM_CHART_DIR ?= $(DATASET_CLONE_DIR)/manifests/dataset
MODEL_NAMESPACE ?= model-system
HF_SECRET_NAME ?= hf-token
HF_SECRET_KEY ?= token
MODELFS_RELEASE ?= modelfs
MODELFS_CHART_DIR ?= charts/modelfs
MODELFS_NAMESPACE ?= modelfs-system
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.28.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq ($(origin GOBIN), undefined)
GOBIN=$(shell go env GOPATH)/bin
else
$(eval GOBIN := $(shell go env GOBIN))
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
CONTAINER_TOOL ?= docker

help:
	@echo "Available targets: generate, manifests, run, tidy, build, docker-build, docker-push, deploy, install"

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
HELM ?= helm

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.18.0

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object paths="./..."

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=charts/modelfs/crds

tidy:
	go fmt ./...

run:
	go run ./main.go

build: generate ## Build manager binary.
	go build -o bin/manager ./main.go

docker-build: ## Build docker image with the manager (for local development).
	$(CONTAINER_TOOL) build -f Dockerfile.dev -t ${IMG} .

docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

##@ Helm Deployment

.PHONY: helm-package
helm-package: manifests ## Package the Helm chart
	$(HELM) package $(MODELFS_CHART_DIR)

.PHONY: helm-install
helm-install: manifests docker-build ## Package and install modelfs using Helm (for local development)
	@echo "Packaging Helm chart..."
	$(HELM) package $(MODELFS_CHART_DIR) --destination /tmp
	@echo "Installing modelfs..."
	@CHART_FILE=$$(ls -t /tmp/modelfs-*.tgz | head -1); \
	$(HELM) upgrade --install $(MODELFS_RELEASE) $$CHART_FILE \
		--namespace $(MODELFS_NAMESPACE) \
		--create-namespace \
		--set image.repository=$(shell echo $(IMG) | cut -d: -f1) \
		--set image.tag=$(shell echo $(IMG) | cut -d: -f2 || echo latest)

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall modelfs Helm release
	$(HELM) uninstall $(MODELFS_RELEASE) --namespace $(MODELFS_NAMESPACE) || true

.PHONY: helm-lint
helm-lint: ## Lint the Helm chart
	$(HELM) lint $(MODELFS_CHART_DIR)

##@ kind + e2e helpers

.PHONY: kind-up
kind-up: ## Create a kind cluster with local-path storage
	kind create cluster --name $(KIND_CLUSTER_NAME) --config $(KIND_CONFIG)
	$(KUBECTL) apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

.PHONY: kind-down
kind-down: ## Delete the kind cluster
	-kind delete cluster --name $(KIND_CLUSTER_NAME)

.PHONY: kind-load-image
kind-load-image: ## Load the controller image into kind
	kind load docker-image $(IMG) --name $(KIND_CLUSTER_NAME)

.PHONY: dataset-clone
dataset-clone:
	@if [ -d "$(DATASET_CLONE_DIR)/.git" ]; then \
	  echo "Dataset submodule already exists, updating..."; \
	  git submodule update --init --recursive $(DATASET_CLONE_DIR); \
	else \
	  echo "Initializing dataset submodule..."; \
	  git submodule update --init --recursive $(DATASET_CLONE_DIR); \
	fi

.PHONY: dataset-install
dataset-install: dataset-clone ## Install BaizeAI/dataset CRDs and Helm chart
	@echo "Installing Dataset CRDs..."
	$(KUBECTL) apply -f $(DATASET_CLONE_DIR)/config/crd/bases/dataset.baizeai.io_datasets.yaml
	@echo "Installing Dataset Helm chart..."
	$(HELM) upgrade --install $(DATASET_RELEASE) $(DATASET_HELM_CHART_DIR) --namespace $(DATASET_NAMESPACE) --create-namespace $(if $(DATASET_HELM_VALUES),-f $(DATASET_HELM_VALUES),)

.PHONY: dataset-uninstall
dataset-uninstall: ## Uninstall BaizeAI/dataset
	-$(HELM) uninstall $(DATASET_RELEASE) --namespace $(DATASET_NAMESPACE)


.PHONY: samples-secret
samples-secret: ## Create HuggingFace token Secret in the model namespace (requires HF_TOKEN env)
	@if [ -z "$(HF_TOKEN)" ]; then echo "HF_TOKEN env var is required" && exit 1; fi
	$(KUBECTL) create namespace $(MODEL_NAMESPACE) --dry-run=client -o yaml | $(KUBECTL) apply -f -
	$(KUBECTL) -n $(MODEL_NAMESPACE) create secret generic $(HF_SECRET_NAME) --from-literal=$(HF_SECRET_KEY)="$(HF_TOKEN)" --dry-run=client -o yaml | $(KUBECTL) apply -f -

.PHONY: samples-apply
samples-apply: ## Apply sample ModelSource + Model manifests to current namespace
	$(KUBECTL) apply -f examples/samples/modelsource-hf.yaml
	$(KUBECTL) apply -f examples/samples/model-qwen.yaml

.PHONY: samples-delete
samples-delete: ## Remove sample resources from current namespace
	-$(KUBECTL) delete -f examples/samples/model-qwen.yaml --ignore-not-found
	-$(KUBECTL) delete -f examples/samples/modelsource-hf.yaml --ignore-not-found

.PHONY: e2e-setup
e2e-setup: kind-up dataset-install docker-build kind-load-image helm-install ## Bootstrap kind, dataset, build+load controller, deploy with Helm

.PHONY: e2e-sample
e2e-sample: samples-apply ## Create sample ModelSource and Model CRs for manual validation
