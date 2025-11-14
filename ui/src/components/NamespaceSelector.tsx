import * as Select from "@radix-ui/react-select";
import { namespaces } from "@mocks/data";
import { useUiState } from "@app/state";

export default function NamespaceSelector() {
  const ns = useUiState((s) => s.namespace);
  const setNs = useUiState((s) => s.setNamespace);
  return (
    <Select.Root value={ns} onValueChange={setNs}>
      <Select.Trigger className="border px-3 py-1 rounded">
        <Select.Value />
      </Select.Trigger>
      <Select.Content className="bg-white border rounded shadow">
        <Select.Viewport>
          {namespaces.map((n) => (
            <Select.Item key={n.name} value={n.name} className="px-3 py-1">
              <Select.ItemText>{n.name}</Select.ItemText>
            </Select.Item>
          ))}
        </Select.Viewport>
      </Select.Content>
    </Select.Root>
  );
}

