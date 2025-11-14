import { render, screen } from "@testing-library/react";
import ModelsPage from "@pages/Models";
import ModelSourcesPage from "@pages/ModelSources";
import ModelDetailPage from "@pages/ModelDetail";
import { vi } from "vitest";

vi.mock("@tanstack/react-router", () => ({ Link: (props: any) => (<a {...props} />) }));

test("Models page basic render", () => {
  render(<ModelsPage />);
  expect(screen.getByText("Models")).toBeInTheDocument();
});

test("ModelSources page basic render", () => {
  render(<ModelSourcesPage />);
  expect(screen.getByText("ModelSources")).toBeInTheDocument();
});

test("ModelDetail page basic loading placeholder", () => {
  Object.defineProperty(window, "location", { value: { pathname: "/models/model-system/qwen3-7b" }, writable: true });
  vi.spyOn(globalThis as any, "fetch").mockResolvedValue({ ok: true, headers: { get: () => "application/json" }, json: async () => ({ summary: { name: "qwen3-7b", namespace: "model-system", sourceRef: "hf-qwen", tags: [], versionsReady: 1, versionsTotal: 2, lastSyncTime: new Date().toISOString(), status: "READY" }, versions: [{ name: "fp16", repo: "qwen/Qwen3-7B", desiredState: "PRESENT", shareEnabled: true, datasetPhase: "READY" }] }) } as any);
  render(<ModelDetailPage />);
  expect(screen.getByText("Loading...")).toBeInTheDocument();
});
