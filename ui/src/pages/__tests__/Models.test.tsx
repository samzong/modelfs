import { render, screen, fireEvent } from "@testing-library/react";
import ModelsPage from "@pages/Models";
import { useUiState } from "@app/state";
import { vi } from "vitest";
vi.mock("@tanstack/react-router", () => ({
  Link: (props: any) => (<a {...props} />),
}));

test("列表筛选按名称或标签", async () => {
  render(<ModelsPage />);
  const input = screen.getByPlaceholderText("筛选名称或标签");
  fireEvent.change(input, { target: { value: "qwen" } });
  expect(useUiState.getState().filterText).toBe("qwen");
});
