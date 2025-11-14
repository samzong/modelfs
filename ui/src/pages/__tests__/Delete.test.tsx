import { render, screen, fireEvent } from "@testing-library/react";
import ModelsPage from "@pages/Models";
import ModelSourcesPage from "@pages/ModelSources";
import { vi } from "vitest";

vi.mock("@tanstack/react-router", () => ({ Link: (props: any) => (<a {...props} />) }));

test("删除模型后不再出现", () => {
  render(<ModelsPage />);
  const delBtn = screen.getAllByText("删除")[0];
  fireEvent.click(delBtn);
  fireEvent.click(screen.getByText("确认删除"));
  expect(screen.queryByText("qwen3-7b")).toBeNull();
});

test("有引用的 ModelSource 删除按钮为禁用", () => {
  render(<ModelSourcesPage />);
  const btns = screen.getAllByText("删除");
  expect(btns[0]).toBeDisabled();
});
