import { render, screen, fireEvent } from "@testing-library/react";
import Wizard from "@pages/ModelWizard";

test("向导 Step1 必填校验", () => {
  render(<Wizard />);
  const nextBtn = screen.getByText("下一步");
  expect(nextBtn).toBeDisabled();
});

test("填写基础信息后进入 Step2 并添加版本", () => {
  render(<Wizard />);
  fireEvent.change(screen.getByLabelText("名称"), { target: { value: "test-model" } });
  fireEvent.change(screen.getByLabelText("命名空间"), { target: { value: "model-system" } });
  fireEvent.change(screen.getByLabelText("来源（SourceRef）"), { target: { value: "hf-qwen" } });
  const nextBtn = screen.getByText("下一步");
  expect(nextBtn).not.toBeDisabled();
  fireEvent.click(nextBtn);
  fireEvent.change(screen.getByPlaceholderText("版本名"), { target: { value: "fp16" } });
  fireEvent.change(screen.getByPlaceholderText("仓库"), { target: { value: "qwen/Qwen3-7B" } });
  const nextBtn2 = screen.getByText("下一步");
  expect(nextBtn2).not.toBeDisabled();
});

