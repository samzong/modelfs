import { render, screen, fireEvent } from "@testing-library/react";
import ModelsPage from "@pages/Models";
import { useUiState } from "@app/state";
import { vi } from "vitest";
vi.mock("@tanstack/react-router", () => ({
  Link: (props: any) => (<a {...props} />),
}));

test("List filter by name or tag", async () => {
  render(<ModelsPage />);
  const input = screen.getByPlaceholderText("Filter by name or tag");
  fireEvent.change(input, { target: { value: "qwen" } });
  expect(useUiState.getState().filterText).toBe("qwen");
});
