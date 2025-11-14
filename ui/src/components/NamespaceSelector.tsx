import * as Select from "@radix-ui/react-select";
import { useUiState } from "@app/state";
import { useDataStore } from "@app/dataStore";
import { useEffect } from "react";

export default function NamespaceSelector() {
  const ns = useUiState((s) => s.namespace);
  const setNs = useUiState((s) => s.setNamespace);
  const { namespaces, refreshNamespaces } = useDataStore();
  useEffect(() => { refreshNamespaces(); }, []);
  useEffect(() => {
    if (namespaces && namespaces.length) {
      const found = namespaces.find((n) => n.name === ns);
      if (!found) {
        setNs(namespaces[0].name);
      }
    }
  }, [namespaces]);
  const items = namespaces && namespaces.length ? namespaces : [{ name: ns }];
  return (
    <Select.Root value={ns} onValueChange={setNs}>
      <Select.Trigger className="border px-3 py-1 rounded">
        <Select.Value placeholder={ns} />
      </Select.Trigger>
      <Select.Content className="bg-white border rounded shadow">
        <Select.Viewport>
          {items.map((n) => (
            <Select.Item key={n.name} value={n.name} className="px-3 py-1">
              <Select.ItemText>{n.name}</Select.ItemText>
            </Select.Item>
          ))}
        </Select.Viewport>
      </Select.Content>
    </Select.Root>
  );
}
