import { render, screen, fireEvent } from "@testing-library/react";
import Wizard from "@pages/ModelWizard";
import { useDataStore } from "@app/dataStore";

test("Wizard Step 1 required validation", () => {
  render(<Wizard />);
  const nextBtn = screen.getByText("Next");
  expect(nextBtn).toBeDisabled();
});

test("Fill basics then go to Step 2 and add version", () => {
  useDataStore.setState({ namespaces: [{ name: "model-system" }], sources: [{ name: "hf-qwen", namespace: "model-system", type: "HUGGING_FACE", secretRef: "", credentialsReady: true, credentialsStatus: "", referencedModels: [], lastChecked: new Date().toISOString() }], models: [], errors: [], loading: false });
  render(<Wizard />);
  fireEvent.change(screen.getByLabelText("Name"), { target: { value: "test-model" } });
  // Namespace and source are ready by default
  const nextBtn = screen.getByText("Next");
  expect(nextBtn).not.toBeDisabled();
  fireEvent.click(nextBtn);
  fireEvent.change(screen.getByPlaceholderText("Version Name"), { target: { value: "fp16" } });
  fireEvent.change(screen.getByPlaceholderText("Repository"), { target: { value: "qwen/Qwen3-7B" } });
  const nextBtn2 = screen.getByText("Next");
  expect(nextBtn2).not.toBeDisabled();
});
