.PHONY: help generate manifests run tidy controller-gen kustomize build docker-build docker-push deploy undeploy install uninstall

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
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
HELM ?= helm

## Tool Versions
KUSTOMIZE_VERSION ?= v5.2.1
CONTROLLER_TOOLS_VERSION ?= v0.18.0

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize && $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object paths="./..."

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

tidy:
	go fmt ./...

run:
	go run ./main.go

build: generate ## Build manager binary.
	go build -o bin/manager ./main.go

docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

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
	mkdir -p $(dir $(DATASET_CLONE_DIR))
	@if [ -d "$(DATASET_CLONE_DIR)/.git" ]; then \
	  git -C $(DATASET_CLONE_DIR) fetch --all --tags --prune && git -C $(DATASET_CLONE_DIR) pull --ff-only; \
	else \
	  git clone $(DATASET_REPO) $(DATASET_CLONE_DIR); \
	fi

.PHONY: dataset-install
dataset-install: dataset-clone ## Install BaizeAI/dataset via bundled Helm chart
	$(HELM) upgrade --install $(DATASET_RELEASE) $(DATASET_HELM_CHART_DIR) --namespace $(DATASET_NAMESPACE) --create-namespace $(if $(DATASET_HELM_VALUES),-f $(DATASET_HELM_VALUES),)

.PHONY: dataset-uninstall
dataset-uninstall: ## Uninstall BaizeAI/dataset
	-$(HELM) uninstall $(DATASET_RELEASE) --namespace $(DATASET_NAMESPACE)

.PHONY: modelfs-deploy-all
modelfs-deploy-all: install deploy ## Install CRDs and deploy controller manifests

.PHONY: modelfs-undeploy-all
modelfs-undeploy-all: undeploy uninstall ## Remove controller and CRDs

.PHONY: samples-secret
samples-secret: ## Create HuggingFace token Secret in the model namespace (requires HF_TOKEN env)
	@if [ -z "$(HF_TOKEN)" ]; then echo "HF_TOKEN env var is required" && exit 1; fi
	$(KUBECTL) create namespace $(MODEL_NAMESPACE) --dry-run=client -o yaml | $(KUBECTL) apply -f -
	$(KUBECTL) -n $(MODEL_NAMESPACE) create secret generic $(HF_SECRET_NAME) --from-literal=$(HF_SECRET_KEY)="$(HF_TOKEN)" --dry-run=client -o yaml | $(KUBECTL) apply -f -

.PHONY: samples-apply
samples-apply: ## Apply sample ModelSource + Model manifests
	$(KUBECTL) apply -f examples/samples/modelsource-hf.yaml
	$(KUBECTL) apply -f examples/samples/model-qwen.yaml

.PHONY: samples-delete
samples-delete: ## Remove sample resources
	-$(KUBECTL) delete -f examples/samples/model-qwen.yaml --ignore-not-found
	-$(KUBECTL) delete -f examples/samples/modelsource-hf.yaml --ignore-not-found

.PHONY: e2e-setup
e2e-setup: kind-up dataset-install docker-build kind-load-image modelfs-deploy-all ## Bootstrap kind, dataset, build+load controller, deploy manifests

.PHONY: e2e-sample
e2e-sample: samples-secret samples-apply ## Create sample Secret/CRs for manual validation
