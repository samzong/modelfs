import { render, screen, fireEvent } from "@testing-library/react";
import Wizard from "@pages/ModelWizard";
import { useDataStore } from "@app/dataStore";

test("向导 Step1 必填校验", () => {
  render(<Wizard />);
  const nextBtn = screen.getByText("下一步");
  expect(nextBtn).toBeDisabled();
});

test("填写基础信息后进入 Step2 并添加版本", () => {
  useDataStore.setState({ namespaces: [{ name: "model-system" }], sources: [{ name: "hf-qwen", namespace: "model-system", type: "HUGGING_FACE", secretRef: "", credentialsReady: true, credentialsStatus: "", referencedModels: [], lastChecked: new Date().toISOString() }], models: [], errors: [], loading: false });
  render(<Wizard />);
  fireEvent.change(screen.getByLabelText("名称"), { target: { value: "test-model" } });
  // 命名空间与来源由默认值自动就绪
  const nextBtn = screen.getByText("下一步");
  expect(nextBtn).not.toBeDisabled();
  fireEvent.click(nextBtn);
  fireEvent.change(screen.getByPlaceholderText("版本名"), { target: { value: "fp16" } });
  fireEvent.change(screen.getByPlaceholderText("仓库"), { target: { value: "qwen/Qwen3-7B" } });
  const nextBtn2 = screen.getByText("下一步");
  expect(nextBtn2).not.toBeDisabled();
});
