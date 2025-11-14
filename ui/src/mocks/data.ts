import type { ModelSummary, ModelDetail, ModelSourceSummary, NamespaceInfo } from "@api/types";

export const namespaces: NamespaceInfo[] = [
  { name: "model-system" },
  { name: "prod" },
  { name: "staging" }
];

export const models: ModelSummary[] = [
  {
    name: "qwen3-7b",
    namespace: "model-system",
    sourceRef: "hf-qwen",
    tags: ["llm", "qwen"],
    versionsReady: 1,
    versionsTotal: 2,
    lastSyncTime: new Date().toISOString(),
    status: "READY"
  },
  {
    name: "llama3-8b",
    namespace: "model-system",
    sourceRef: "hf-llama3",
    tags: ["llm", "meta"],
    versionsReady: 0,
    versionsTotal: 1,
    lastSyncTime: new Date().toISOString(),
    status: "PROCESSING",
    reconcileMessage: "Syncing weights"
  }
];

export const modelDetails: Record<string, ModelDetail> = {
  "model-system/qwen3-7b": {
    summary: models[0],
    description: "Qwen3 7B base model",
    versions: [
      {
        name: "fp16",
        repo: "qwen/Qwen3-7B",
        desiredState: "PRESENT",
        shareEnabled: true,
        datasetPhase: "READY",
        pvcName: "mdl-qwen3-7b-fp16",
      },
      {
        name: "int4",
        repo: "qwen/Qwen3-7B",
        desiredState: "PRESENT",
        shareEnabled: false,
        datasetPhase: "PENDING",
      }
    ]
  }
};

export const modelSources: ModelSourceSummary[] = [
  {
    name: "hf-qwen",
    namespace: "model-system",
    type: "HUGGING_FACE",
    secretRef: "hf-token",
    credentialsReady: true,
    credentialsStatus: "OK",
    referencedModels: ["model-system/qwen3-7b"],
    lastChecked: new Date().toISOString(),
  }
];

