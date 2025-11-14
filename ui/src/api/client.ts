import { apiFetch } from "@api/http";
import type { ModelSummary, ModelDetail, ModelSourceSummary, NamespaceInfo, ErrorBanner } from "@api/types";

export const client = {
  async listModels(ns: string): Promise<{ items: ModelSummary[] }> {
    return apiFetch(`/api/models?namespace=${encodeURIComponent(ns)}`);
  },
  async getModel(ns: string, name: string): Promise<ModelDetail> {
    return apiFetch(`/api/models/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`);
  },
  async createModel(ns: string, req: { name: string; sourceRef: string; description?: string; tags?: string[]; versions: Array<{ name: string; repo: string; revision?: string; precision?: string; desiredState?: string; shareEnabled?: boolean; }> }): Promise<ModelDetail> {
    return apiFetch(`/api/models`, { method: "POST", body: JSON.stringify({ ...req, namespace: ns }) });
  },
  async updateModel(ns: string, name: string, req: { sourceRef: string; description?: string; tags?: string[]; versions: Array<{ name: string; repo: string; revision?: string; precision?: string; desiredState?: string; shareEnabled?: boolean; }> }): Promise<ModelDetail> {
    return apiFetch(`/api/models/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`, { method: "PUT", body: JSON.stringify({ name, namespace: ns, ...req }) });
  },
  async listModelSources(ns: string): Promise<{ items: ModelSourceSummary[] }> {
    return apiFetch(`/api/modelsources?namespace=${encodeURIComponent(ns)}`);
  },
  async getModelSource(ns: string, name: string): Promise<{ name: string; namespace: string; spec: any }> {
    return apiFetch(`/api/modelsources/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`);
  },
  async updateModelSource(ns: string, name: string, req: { type: string; secretRef?: string; config?: Record<string, string> }): Promise<void> {
    await apiFetch(`/api/modelsources/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`, { method: "PUT", body: JSON.stringify({ name, namespace: ns, ...req }) });
  },
  async deleteModelSource(ns: string, name: string): Promise<void> {
    await apiFetch(`/api/modelsources/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`, { method: "DELETE" });
  },
  async listNamespaces(): Promise<{ items: NamespaceInfo[] }> {
    return apiFetch(`/api/namespaces`);
  },
  async listErrors(ns: string): Promise<{ items: ErrorBanner[] }> {
    return apiFetch(`/api/errors?namespace=${encodeURIComponent(ns)}`);
  },
  async createModelSource(ns: string, req: { name: string; type: string; secretRef?: string; config?: Record<string, string> }): Promise<void> {
    await apiFetch(`/api/modelsources`, { method: "POST", body: JSON.stringify({ ...req, namespace: ns }) });
  },
  async validateSecret(ns: string, name: string): Promise<{ ready: boolean; message: string }> {
    return apiFetch(`/api/secrets/validate?namespace=${encodeURIComponent(ns)}&name=${encodeURIComponent(name)}`);
  },
  async listDatasets(ns: string): Promise<{ items: Array<{ name: string; namespace: string; phase: string; pvcName?: string; lastSync: string }> }> {
    return apiFetch(`/api/datasets?namespace=${encodeURIComponent(ns)}`);
  },
  async deleteModel(ns: string, name: string): Promise<void> {
    await apiFetch(`/api/models/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`, { method: "DELETE" });
  },
  async deleteModelVersion(ns: string, name: string, version: string): Promise<void> {
    await apiFetch(`/api/models/${encodeURIComponent(ns)}/${encodeURIComponent(name)}/versions/${encodeURIComponent(version)}`, { method: "DELETE" });
  },
  async toggleShare(ns: string, name: string, version: string, enabled: boolean): Promise<void> {
    await apiFetch(`/api/models/${encodeURIComponent(ns)}/${encodeURIComponent(name)}/versions/${encodeURIComponent(version)}/share`, { method: "POST", body: JSON.stringify({ enabled }) });
  },
  async triggerResync(ns: string, name: string): Promise<void> {
    await apiFetch(`/api/models/${encodeURIComponent(ns)}/${encodeURIComponent(name)}/actions/resync`, { method: "POST" });
  },
};
