import { render, screen, fireEvent } from "@testing-library/react";
import ModelsPage from "@pages/Models";
import ModelSourcesPage from "@pages/ModelSources";
import { vi } from "vitest";
import { useDataStore } from "@app/dataStore";

vi.mock("@tanstack/react-router", () => ({ Link: (props: any) => (<a {...props} />) }));

test("删除模型后不再出现", () => {
  useDataStore.setState({ models: [{ name: "qwen3-7b", namespace: "model-system", sourceRef: "hf-qwen", tags: [], versionsReady: 1, versionsTotal: 2, lastSyncTime: new Date().toISOString(), status: "READY" }], sources: [], errors: [], loading: false });
  render(<ModelsPage />);
  const delBtn = screen.getAllByText("删除")[0];
  fireEvent.click(delBtn);
  fireEvent.click(screen.getByText("确认删除"));
  expect(screen.queryByText("qwen3-7b")).toBeNull();
});

test("有引用的 ModelSource 删除按钮为禁用", () => {
  useDataStore.setState({ sources: [{ name: "hf-qwen", namespace: "model-system", type: "HUGGING_FACE", secretRef: "hf-token", credentialsReady: true, credentialsStatus: "OK", referencedModels: ["model-system/qwen3-7b"], lastChecked: new Date().toISOString() }], models: [], errors: [], loading: false });
  render(<ModelSourcesPage />);
  const btns = screen.getAllByText("删除");
  expect(btns[0]).toBeDisabled();
});
