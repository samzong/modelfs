import { create } from "zustand";
import type { ModelSummary, ModelSourceSummary, ErrorBanner, NamespaceInfo } from "@api/types";
import { client } from "@api/client";
import { attachSSE } from "@api/sse";

type State = {
  models: ModelSummary[];
  sources: ModelSourceSummary[];
  errors: ErrorBanner[];
  namespaces: NamespaceInfo[];
  datasets: Array<{ name: string; namespace: string; phase: string; pvcName?: string; lastSync: string }>;
  loading: boolean;
  refreshAll: (ns: string) => Promise<void>;
  refreshNamespaces: () => Promise<void>;
  attachSSE: (ns: string) => () => void;
  removeModelLocal: (ns: string, name: string) => void;
};

export const useDataStore = create<State>((set, get) => ({
  models: [],
  sources: [],
  errors: [],
  namespaces: [],
  datasets: [],
  loading: false,
  refreshAll: async (ns: string) => {
    set({ loading: true });
    try {
      const [ml, src, err, ds] = await Promise.all([
        client.listModels(ns),
        client.listModelSources(ns),
        client.listErrors(ns),
        client.listDatasets(ns),
      ]);
      set({ models: ml?.items ?? [], sources: src?.items ?? [], errors: err?.items ?? [], datasets: ds?.items ?? [] });
    } catch {
    } finally {
      set({ loading: false });
    }
  },
  refreshNamespaces: async () => {
    try {
      const ns = await client.listNamespaces();
      set({ namespaces: ns.items });
    } catch {}
  },
  attachSSE: (ns: string) => attachSSE(ns, (resource, action, payload) => {
    const st = get();
    if (resource === "models") {
      const items = st.models.slice();
      const idx = items.findIndex(m => m.namespace === payload.namespace && m.name === payload.name);
      if (action === "deleted" && idx >= 0) items.splice(idx, 1);
      else if (idx >= 0) items[idx] = payload;
      else if (action === "added") items.unshift(payload);
      set({ models: items });
    } else if (resource === "modelsources") {
      const items = st.sources.slice();
      const idx = items.findIndex(s => s.namespace === payload.namespace && s.name === payload.name);
      if (action === "deleted" && idx >= 0) items.splice(idx, 1);
      else if (idx >= 0) items[idx] = payload;
      else if (action === "added") items.unshift(payload);
      set({ sources: items });
    } else if (resource === "datasets") {
      const items = st.datasets.slice();
      const idx = items.findIndex(d => d.namespace === payload.namespace && d.name === payload.name);
      if (action === "deleted" && idx >= 0) items.splice(idx, 1);
      else if (idx >= 0) items[idx] = payload;
      else if (action === "added") items.unshift(payload);
      set({ datasets: items });
    }
  }),
  removeModelLocal: (ns: string, name: string) => {
    const items = get().models.filter(m => !(m.namespace === ns && m.name === name));
    set({ models: items });
  },
}));
