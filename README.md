# modelfs

`modelfs` is a Kubernetes operator for declarative llm model weight management, built on [BaizeAI/dataset](https://github.com/BaizeAI/dataset).

## Core Concepts

### ModelSource

Defines connection information and common configurations (e.g., authentication, file filters) for a model source. Can be reused by multiple models.

**Supported source types**: `HUGGING_FACE`, `MODEL_SCOPE`, `S3`, `HTTP`, `GIT`, `PVC`, `NFS`, `CONDA`, `REFERENCE`

### Model

Defines a model with multiple versions. Each version has its own repository path, storage configuration, and sharing settings.

**Key fields**:

- `sourceRef`: References a `ModelSource` for connection/auth info
- `versions`: List of model versions, each with:
  - `name`: Version identifier
  - `repo`: Repository path (e.g., `qwen/Qwen2.5-7B-Instruct`)
  - `revision`: Git revision (default: `main`)
  - `storage`: PVC configuration (access modes, size, storage class)
  - `state`: `PRESENT` (sync) or `ABSENT` (delete)
  - `share`: Cross-namespace sharing configuration

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
     versions:
       - name: v2.5.0
         repo: qwen/Qwen2.5-7B-Instruct
         revision: main
         precision: FP16
         storage:
           accessModes:
             - ReadWriteOnce
           resources:
             requests:
               storage: 50Gi
           storageClassName: local-path
         state: PRESENT
   ```

## Architecture

```
User creates:
  ModelSource (connection/auth) → Model (versions + repos)
                                         ↓
modelfs Controller creates:      Dataset CR per version
                                         ↓
BaizeAI/dataset Controller:      Downloads weights
                                         ↓
                                   PVC ready for use
```

## Integration with BaizeAI/dataset

- `modelfs` controllers directly create and manage `Dataset` CRs via Kubernetes API
- Each `Model` version creates a `Dataset` CR named `mdl-{model-name}-{version-name}`
- `Model` status aggregates `Dataset` status (phase, conditions, PVC name, last sync time)
- Dataset reconciliation is handled by BaizeAI/dataset controllers
- Cross-namespace sharing uses REFERENCE Dataset type

## Project Structure

- `api/v1/`: CRD type definitions (`Model`, `ModelSource`)
- `controllers/`: Reconciliation logic for Model and ModelSource CRDs
- `pkg/dataset/`: Client for creating/managing `Dataset` CRs
- `config/`: Kubernetes manifests (CRDs, RBAC, samples)
- `examples/`: Sample manifests for common use cases
