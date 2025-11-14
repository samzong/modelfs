import { modelSources } from "@mocks/data";
import { useUiState } from "@app/state";
import Card from "@components/Card";
import Badge from "@components/Badge";
import SectionHeader from "@components/SectionHeader";
import Button from "@components/Button";
import { useState } from "react";
import ConfirmDialog from "@components/ConfirmDialog";

export default function ModelSourcesPage() {
  const ns = useUiState((s) => s.namespace);
  const [refresh, setRefresh] = useState(0);
  const rows = modelSources.filter((s) => s.namespace === ns);
  const [pendingDelete, setPendingDelete] = useState<string | null>(null);
  function onDelete(name: string) { setPendingDelete(name); }
  function confirmDelete() {
    if (!pendingDelete) return;
    const idx = modelSources.findIndex(s => s.namespace === ns && s.name === pendingDelete);
    if (idx >= 0 && !(modelSources[idx].referencedModels && modelSources[idx].referencedModels!.length > 0)) {
      modelSources.splice(idx, 1);
      setRefresh(refresh + 1);
    }
    setPendingDelete(null);
  }
  function cancelDelete() { setPendingDelete(null); }
  return (
    <div className="space-y-4">
      <SectionHeader title="ModelSources" description="模型来源配置与凭据状态" />
      <Card>
        <table className="min-w-full">
          <thead>
            <tr className="text-left text-sm text-gray-500">
              <th className="p-2">名称</th>
              <th className="p-2">类型</th>
              <th className="p-2">Secret</th>
              <th className="p-2">凭据</th>
              <th className="p-2">引用模型</th>
              <th className="p-2">操作</th>
            </tr>
          </thead>
          <tbody>
            {rows.length === 0 ? (
              <tr><td className="p-6 text-gray-500" colSpan={5}>暂无来源</td></tr>
            ) : rows.map((s, idx) => (
              <tr key={`${s.namespace}/${s.name}`} className={idx % 2 === 0 ? "bg-white" : "bg-muted"}>
                <td className="p-2">{s.name}</td>
                <td className="p-2">{s.type}</td>
                <td className="p-2">{s.secretRef || "-"}</td>
                <td className="p-2">{s.credentialsReady ? <span className="px-2 py-1 rounded-lg bg-green-100 text-green-800 text-sm">就绪</span> : s.credentialsStatus || "未知"}</td>
                <td className="p-2">{s.referencedModels?.length ? s.referencedModels?.join(", ") : "-"}</td>
                <td className="p-2">
                  <Button size="sm" variant="danger" onClick={() => onDelete(s.name)} disabled={!!(s.referencedModels && s.referencedModels.length > 0)}>删除</Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
      <ConfirmDialog open={!!pendingDelete} title={`删除 ModelSource ${pendingDelete || ""}`} description="当存在引用模型时不可删除（Mock）。" onConfirm={confirmDelete} onCancel={cancelDelete} />
    </div>
  );
}
