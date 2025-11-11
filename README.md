# modelfs

`modelfs` is a Kubernetes operator for declarative llm model weight management, built on [BaizeAI/dataset](https://github.com/BaizeAI/dataset).

## Core Concepts

### ModelSource

Defines connection information and common configurations (e.g., authentication, file filters) for a model source. Can be reused by multiple models.

**Supported source types**: `HUGGING_FACE`, `MODEL_SCOPE`, `S3`, `HTTP`, `GIT`, `PVC`, `NFS`, `CONDA`, `REFERENCE`

### Model

Defines a model with multiple versions. Each version has its own repository path and configuration. The version list is derived from the keys of `versionConfigs`.

**Key fields**:

- `sourceRef`: References a `ModelSource` for connection/auth info
- `versionConfigs`: Map of version names to version-specific configs (repo, repoConfig)

### ModelSync

Triggers synchronization for a specific version of a model. Creates and manages `Dataset` CRs through BaizeAI/dataset.

**Key fields**:

- `modelRef`: References a `Model` (ModelSource is resolved from `Model.sourceRef`)
- `version`: Version to sync (must exist in `Model.versionConfigs`)

### ModelReference

Enables cross-namespace sharing of cached models as read-only references.

## Quick Start

1. **Install BaizeAI/dataset** (required dependency):

   ```bash
   # Follow BaizeAI/dataset installation instructions
   ```

2. **Install modelfs**:

   ```bash
   kubectl apply -k config/default
   ```

3. **Create a ModelSource**:

   ```yaml
   apiVersion: model.samzong.dev/v1
   kind: ModelSource
   metadata:
     name: huggingface-source
   spec:
     type: HUGGING_FACE
     config:
       endpoint: https://huggingface.co
       include: "*.safetensors,*.json"
       exclude: "*.bin"
   ```

4. **Create a Model**:

   ```yaml
   apiVersion: model.samzong.dev/v1
   kind: Model
   metadata:
     name: qwen-model
   spec:
     sourceRef: huggingface-source
     versionConfigs:
       v2.5.0:
         repo: qwen/Qwen2.5-7B-Instruct
         repoConfig:
           revision: main
   ```

5. **Create a ModelSync**:
   ```yaml
   apiVersion: model.samzong.dev/v1
   kind: ModelSync
   metadata:
     name: qwen-sync
   spec:
     modelRef: qwen-model
     version: v2.5.0
   ```

## Architecture

```
User creates:
  ModelSource (connection/auth) → Model (versions + repos) → ModelSync (trigger sync)
                                                                    ↓
modelfs Controller creates:                                    Dataset CR
                                                                    ↓
BaizeAI/dataset Controller:                                 Downloads weights
                                                                    ↓
                                                              PVC ready for use
```

## Integration with BaizeAI/dataset

- `modelfs` controllers directly create and manage `Dataset` CRs via Kubernetes API
- Each `ModelSync` creates a `Dataset` CR named `{model-name}-{version}`
- `ModelSync` status mirrors `Dataset` status (phase, conditions, last sync time)
- Dataset reconciliation is handled by BaizeAI/dataset controllers

## Project Structure

- `api/v1/`: CRD type definitions (`Model`, `ModelSource`, `ModelSync`, `ModelReference`)
- `controllers/`: Reconciliation logic for each CRD
- `pkg/dataset/`: Client for creating/managing `Dataset` CRs
- `config/`: Kubernetes manifests (CRDs, RBAC, samples)
