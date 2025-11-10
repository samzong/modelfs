# modelfs

`modelfs` is a Kubernetes-native model weight management component built on top of [BaizeAI/dataset](https://github.com/BaizeAI/dataset). It focuses on declaratively publishing, synchronizing, and reusing model weights across the cluster.

## Core capabilities
- **Declarative model lifecycle**: custom resources `Model`, `ModelSource`, and `ModelSync` capture the model definition, provenance, and sync schedules.
- **Data-plane reuse**: relies on BaizeAI/dataset for data loading and PVC warming to pull external model weights into persistent volumes inside the cluster.
- **Multi-source sync**: one workflow handles Hugging Face, S3, HTTP, NFS, and other endpoints.
- **Cross-namespace sharing**: `ModelReference` exposes cached models as read-only to other namespaces.

## Resources and controller responsibilities
| Resource | Controller responsibility | Notes |
|----------|---------------------------|-------|
| `Model` | Maintain model metadata, track versions, and default sources | Anchor for other resources |
| `ModelSource` | Validate and wrap external source configuration | Supports multiple protocols and auth schemes |
| `ModelSync` | Schedule weight synchronization jobs and coordinate PVC mounts | Triggers the dataset sync pipeline |
| `ModelReference` | Map models to target namespaces | Governs access scope and read-only policy |

## Architecture overview
1. Operators declare models, sources, and sync policies via CRDs.
2. Controllers watch resource changes and call BaizeAI/dataset to fetch and cache weights.
3. Once synced, weights are surfaced through PVCs for inference or training workloads.
4. Optional `ModelReference` resources let other namespaces reuse cached models.

## Code layout
> Directory names follow Kubebuilder project conventions.
- `api/`: CRD structs and schema definitions.
- `controllers/`: Reconciliation logic for each resource, bridging to the dataset sync.
- `config/`: Kubernetes manifests for CRDs, RBAC, webhooks, and samples.
- `pkg/`: Reusable integrations with the dataset project and storage backends.

These directories will be populated with concrete controllers and docs over time.
