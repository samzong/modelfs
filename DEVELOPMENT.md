# Development Guide

## Prerequisites

- Go 1.25+, Docker, kubectl, kind, helm, Make
- Git (for submodule support)

## Initial Setup

```bash
# Clone repository with submodules
git clone --recursive https://github.com/samzong/modelfs.git
cd modelfs

# Or if already cloned, initialize submodules
git submodule update --init --recursive
```

## Development Workflow

```bash
# 1. Create cluster
make kind-up

# 2. Install dataset (includes CRD installation)
# This will:
#   - Initialize/update the dataset submodule
#   - Install Dataset CRDs
#   - Install Dataset Helm chart
make dataset-install

# 3. Build and deploy modelfs using Helm
make docker-build
make kind-load-image
make helm-install

# Or install directly using helm command:
# helm install modelfs ./charts/modelfs --namespace modelfs-system --create-namespace

# 4. Deploy sample to current namespace (no token required for public models)
# Resources will be deployed to the namespace set in your kubectl context
make e2e-sample
```

## Updating Dataset Submodule

```bash
# Update dataset submodule to latest version
git submodule update --remote third_party/dataset

# Or update to a specific tag/commit
cd third_party/dataset
git checkout v0.1.6  # or specific commit
cd ../..
git add third_party/dataset
git commit -m "Update dataset submodule"
```

## Verify

```bash
# Check modelfs controller
kubectl -n modelfs-system get pods

# Check CRDs
kubectl get crd | grep -E "(model|dataset)"

# Check resources (in current namespace or specify namespace)
kubectl get model,modelsource,dataset
# Or specify namespace: kubectl -n <namespace> get model,modelsource,dataset
```

## Cleanup

```bash
# Remove sample resources
make samples-delete

# Remove modelfs controller and CRDs
make helm-uninstall

# Remove dataset (Helm chart only, CRDs remain)
make dataset-uninstall

# To remove Dataset CRDs manually:
kubectl delete crd datasets.dataset.baizeai.io

# Delete cluster
make kind-down
```

## Helm Chart Development

```bash
# Lint the Helm chart
make helm-lint

# Package the chart
make helm-package

# Install locally (packages and installs)
make helm-install

# Uninstall
make helm-uninstall
```

## Notes

- **Dataset Submodule**: The `third_party/dataset` directory is a git submodule pointing to BaizeAI/dataset repository. It's used only for local development and debugging (e.g., `make dataset-install`). Always use `git submodule` commands to update it.
- **CRD Installation**: Dataset CRDs are installed separately via `make dataset-install` (not included in modelfs CRDs).
- **Submodule Updates**: When pulling changes, remember to run `git submodule update --init --recursive` if submodules are updated.
- **ModelSource SecretRef**: The `secretRef` field in ModelSource is optional. For public models (e.g., HuggingFace public repos), you can omit it. Only set `secretRef` when accessing private repositories or when authentication is required.
- **Helm Deployment**: modelfs is deployed using Helm chart. The chart is located in `charts/modelfs/` directory. CRDs are generated to `charts/modelfs/crds/` when running `make manifests`.
