# modelfs Helm Chart

A Helm chart for deploying the modelfs Kubernetes operator.

## Introduction

This chart deploys the modelfs controller manager on a Kubernetes cluster using the Helm package manager.

## Prerequisites

- Kubernetes 1.28+
- Helm 3.0+
- BaizeAI/dataset operator installed (CRDs and controller)

## Installing the Chart

To install the chart with the release name `modelfs`:

```bash
helm install modelfs ./charts/modelfs --namespace modelfs-system --create-namespace
```

To install with custom values:

```bash
helm install modelfs ./charts/modelfs --namespace modelfs-system --create-namespace -f my-values.yaml
```

## Uninstalling the Chart

To uninstall/delete the `modelfs` deployment:

```bash
helm uninstall modelfs
```

## Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image repository | `controller` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `namespace.create` | Create namespace | `true` |
| `namespace.name` | Namespace name | `modelfs-system` |
| `replicaCount` | Number of replicas | `1` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `""` |
| `rbac.create` | Create RBAC resources | `true` |
| `leaderElection.enabled` | Enable leader election | `true` |
| `leaderElection.resourceName` | Leader election resource name | `modelfs-leader-election` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `10m` |
| `resources.requests.memory` | Memory request | `64Mi` |

## Examples

### Install with custom image

```bash
helm install modelfs ./charts/modelfs \
  --set image.repository=my-registry/modelfs-controller \
  --set image.tag=v0.1.0
```

### Install without creating namespace

```bash
helm install modelfs ./charts/modelfs \
  --set namespace.create=false \
  --set namespace.name=existing-namespace
```

### Install with custom resources

```bash
helm install modelfs ./charts/modelfs \
  --set resources.limits.cpu=1000m \
  --set resources.limits.memory=1Gi \
  --set resources.requests.cpu=100m \
  --set resources.requests.memory=128Mi
```

## Upgrading

To upgrade the chart:

```bash
helm upgrade modelfs ./charts/modelfs
```

To upgrade with custom values:

```bash
helm upgrade modelfs ./charts/modelfs -f my-values.yaml
```

## Troubleshooting

### Check controller status

```bash
kubectl get pods -n modelfs-system
kubectl logs -n modelfs-system -l control-plane=controller-manager
```

### Check CRDs

```bash
kubectl get crd | grep model.samzong.dev
```

## Support

For issues and questions, please visit: https://github.com/samzong/modelfs

