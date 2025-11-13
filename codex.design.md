# modelfs UI MVP Design

## Objective & Scope

Deliver a thin web console that surfaces the state of `Model` and `ModelSource` CRDs (see `api/v1/model_types.go` and `api/v1/modelsource_types.go`) and enables basic CRUD without forcing operators back to `kubectl`. Every interaction is scoped to a single namespace selected upfront—no cross-namespace dashboards. The MVP must cover:

1. Read-only visibility into inventory, health, and sync progress for models/versions.
2. Guided creation and editing of ModelSources and Models (including versions, storage, and sharing fields).
3. Safe operations for day‑2 management: delete model/version, re-run sync, inspect underlying Dataset/PVC linkage.
   Out of scope for MVP: authN/Z (reuse cluster RBAC), custom dashboards, Dataset-level tuning, or multi-cluster federation.

## Primary Users & Stories

- **MLOps Engineer (namespace admin).** Needs to check whether model weights finished syncing, roll out a new version, or pause a broken revision.
  - View all models within the currently selected namespace, filter by tags, inspect per-version phases (`controllers/model_controller.go`).
  - Launch a “Create Model” flow that references an existing ModelSource and defines multiple versions.
- **Platform SRE (cluster admin).** Owns ModelSources and secrets, watches for failures, and ensures sharing policies are honored.
  - Validate credential readiness for each ModelSource (condition `CredentialsReady` from `controllers/modelsource_controller.go`).
  - Audit which namespaces received shared datasets and delete stale references.

## Architecture Overview

- **UI**: React + Vite SPA served via `modelfs-ui` pod; communicates with backend via REST + SSE for watch streams.
- **Backend Gateway**: Go service running in-cluster, using controller-runtime client to talk to Kubernetes API. Provides:
  - `/api/models`, `/api/models/{ns}/{name}` (CRUD + watch endpoints).
  - `/api/modelsources`, `/api/secrets/validate`, `/api/datasets` (read-only for Dataset phase lookup from `dataset.baizeai.io`).
  - Auth relies on the pod's ServiceAccount token, and the UI keeps a short-lived in-cluster session cookie so the browser never handles kube credentials.
- **Event pipeline**: backend watches Models, ModelSources, and Datasets within the selected namespace (mirroring controller watch set) and pushes diff-friendly payloads over SSE to keep tables live.

## Information Architecture

1. **Global Layout**: left nav with just two primary entries—Models and ModelSources. The top bar contains a kube-context indicator and a mandatory namespace dropdown; users must switch namespaces before seeing other data.
2. **Models List**
   - Columns: Name, Namespace, Source, Versions Ready/Total, Last Sync (latest `SyncedVersion.LastSyncTime`), Status pill (aggregated from `SyncedVersion.Phase`).
   - Filters: tag, phase, source type (namespace filter is redundant because of forced context).
   - Row actions: view, edit YAML, delete.
3. **Model Detail**
   - Header with description, tags, SourceRef, sharing summary.
   - **Versions table**: each row shows Repo@Revision, Precision, Storage request, Share status, Dataset Phase, PVC, Observed hash. Clicking opens side panel with Dataset events, conditions, and `kubectl` snippet.
   - **YAML/spec tab**: read-only spec diff with copy button.
   - Actions: edit spec (launch wizard pre-filled), add version, delete version, toggle share.
4. **Model Wizard (Create/Edit)**
   - Step 1: Basics (name, namespace, tags, display info, SourceRef autocomplete from ModelSources).
   - Step 2: Versions (repeatable cards). Each card includes Repo, Revision, Precision (enum `FP16|INT4|INT8`), storage (PVC size/class), desired state, share config (namespace selector YAML editor + opt-in label string) per `model_types.go`.
   - Step 3: Review & YAML preview, plus CLI equivalent (`kubectl apply -f -`).
5. **ModelSource List & Detail**
   - Columns: Name, Namespace, Type, SecretRef, Credentials status, Referenced Models count.
   - Detail shows config key/values, referencing models list, and secret validation status with timestamp.
   - Actions: edit config, rotate secret (links to instructions), delete with safeguard if `status.referencedBy` non-empty.

## Data & Interaction Notes

- **Status aggregation**: compute Model row status as worst of `SyncedVersion.Phase` (e.g., Failed > Processing > Pending > Ready). Provide explanation tooltip referencing dataset names `mdl-<model>-<version>` generated in `pkg/dataset/client.go`.
- **Sharing insights**: show the namespaces matched by `share.namespaceSelector` and `requireOptInLabel`. Query namespaces once and cache; highlight when selector yields zero matches.
- **Storage display**: show requested vs observed storage (read from `SyncedVersion.ObservedStorage`). Provide “view PVC” link (`kubectl -n <ns> get pvc <name>` snippet).
- **Secret validation**: backend performs dry-run GET on referenced Secret to mirror logic in `ModelSourceReconciler.validateSecret`.
- **Error surfacing**: when controller writes `ReconcileError`, surface inline banner above the Models table with reason + retry ETA. Offer CTA to download logs (`kubectl logs deploy/modelfs-controller -n modelfs-system`).
- **Edits**: forms post JSON spec fragments; backend converts to CR YAML and applies via server-side apply to preserve status.

## Technology Choices

- Frontend: React 18 + TanStack Router/Table, Tailwind for fast layout, Radix UI components (cards, tabs, dialogs).
- Backend: Go 1.25 (matching repo), controller-runtime client for typed CR access, gorilla/mux or chi for HTTP, `sse` package for watch streaming. Same repo module (`cmd/ui-server/main.go`) ensures version lockstep.
- Packaging: extend Helm chart (`charts/modelfs/`) with optional `ui.enabled` flag deploying the UI service + ingress.

## RBAC & Next Steps

- RBAC: ship a single `modelfs-ui` ServiceAccount bound to a ClusterRole that can manage Model/ModelSource CRDs cluster-wide and read supporting objects (`datasets.dataset.baizeai.io`, `secrets`, `namespaces`, `persistentvolumeclaims`). Favor simplicity over fine-grained scoping for the MVP.

**Immediate follow-ups**

- [ ] Prototype backend skeleton under `cmd/ui-server/` using existing module deps.
- [ ] Extend Helm chart with UI deployment + service account + ClusterRole.
- [ ] Build Models list + detail views with mock data to validate UX before wiring live watches.
