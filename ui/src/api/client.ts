import { apiFetch } from "@api/http";
import type { ModelSummary, ModelDetail, ModelSourceSummary, NamespaceInfo, ErrorBanner } from "@api/types";

export const client = {
  async listModels(ns: string): Promise<{ items: ModelSummary[] }> {
    return apiFetch(`/api/models?namespace=${encodeURIComponent(ns)}`);
  },
  async getModel(ns: string, name: string): Promise<ModelDetail> {
    return apiFetch(`/api/models/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`);
  },
  async listModelSources(ns: string): Promise<{ items: ModelSourceSummary[] }> {
    return apiFetch(`/api/modelsources?namespace=${encodeURIComponent(ns)}`);
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
