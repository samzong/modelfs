import { create } from "zustand";

interface UiState {
  namespace: string;
  setNamespace: (ns: string) => void;
  filterText: string;
  setFilterText: (v: string) => void;
}

export const useUiState = create<UiState>((set) => ({
  namespace: "model-system",
  setNamespace: (ns) => set({ namespace: ns }),
  filterText: "",
  setFilterText: (v) => set({ filterText: v }),
}));

