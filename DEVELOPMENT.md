# Development Guide

## Prerequisites

- Go 1.25+, Docker, kubectl, kind, helm, Make

## Setup

```bash
# 1. Create cluster
make kind-up

# 2. Install dataset
make dataset-install

# 3. Build and deploy modelfs
make docker-build
make kind-load-image
make modelfs-deploy-all

# 4. Deploy sample
export HF_TOKEN=your_token
make e2e-sample
```

## Verify

```bash
kubectl -n modelfs-system get pods
kubectl -n model-system get model,modelsource,dataset
```

## Cleanup

```bash
make samples-delete
make modelfs-undeploy-all
make dataset-uninstall
make kind-down
```
