import { create } from "zustand";

interface UiState {
  namespace: string;
  setNamespace: (ns: string) => void;
  filterText: string;
  setFilterText: (v: string) => void;
}

export const useUiState = create<UiState>((set) => ({
  namespace: (typeof window !== "undefined" && window.localStorage.getItem("modelfs_ns")) || "model-system",
  setNamespace: (ns) => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem("modelfs_ns", ns);
    }
    set({ namespace: ns });
  },
  filterText: "",
  setFilterText: (v) => set({ filterText: v }),
}));
