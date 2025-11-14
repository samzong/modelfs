import { render, screen } from "@testing-library/react";
import ModelsPage from "@pages/Models";
import ModelSourcesPage from "@pages/ModelSources";
import ModelDetailPage from "@pages/ModelDetail";
import { vi } from "vitest";

vi.mock("@tanstack/react-router", () => ({ Link: (props: any) => (<a {...props} />) }));

test("Models 页面基础渲染", () => {
  render(<ModelsPage />);
  expect(screen.getByText("Models")).toBeInTheDocument();
});

test("ModelSources 页面基础渲染", () => {
  render(<ModelSourcesPage />);
  expect(screen.getByText("ModelSources")).toBeInTheDocument();
});

test("ModelDetail 页面基础渲染占位", () => {
  Object.defineProperty(window, "location", { value: { pathname: "/models/model-system/qwen3-7b" }, writable: true });
  render(<ModelDetailPage />);
  expect(screen.getAllByText("版本").length).toBeGreaterThan(0);
});
