import { useMemo, useState, useEffect } from "react";
import Badge from "@components/Badge";
import Button from "@components/Button";
import Card from "@components/Card";
import SectionHeader from "@components/SectionHeader";
import { useUiState } from "@app/state";
import { Link } from "@tanstack/react-router";
import ConfirmDialog from "@components/ConfirmDialog";
import { useDataStore } from "@app/dataStore";
import ErrorBanner from "@components/ErrorBanner";
import { client } from "@api/client";

export default function ModelsPage() {
  const filterText = useUiState((s) => s.filterText);
  const setFilterText = useUiState((s) => s.setFilterText);
  const ns = useUiState((s) => s.namespace);
  const [pendingDelete, setPendingDelete] = useState<string | null>(null);
  const { models, errors, refreshAll, attachSSE, removeModelLocal } = useDataStore();
  useEffect(() => {
    refreshAll(ns);
    const close = attachSSE(ns);
    return () => close();
  }, [ns]);
  const filtered = useMemo(() => models.filter((m) => m.namespace === ns && (!filterText || m.name.includes(filterText) || m.tags?.some((t) => t.includes(filterText)))), [ns, filterText, models]);
  function onDelete(name: string) { setPendingDelete(name); }
  function confirmDelete() {
    if (!pendingDelete) return;
    client.deleteModel(ns, pendingDelete).catch(() => {});
    removeModelLocal(ns, pendingDelete);
    setPendingDelete(null);
  }
  function cancelDelete() { setPendingDelete(null); }
  return (
    <div className="space-y-4">
      <SectionHeader title="Models" description="命名空间内的模型清单" right={<Link to="/models/wizard"><Button variant="primary">创建</Button></Link>} />
      <div className="toolbar">
        <input className="form-input w-72" placeholder="筛选名称或标签" value={filterText} onChange={(e) => setFilterText(e.target.value)} />
      </div>
      <ErrorBanner items={errors} />
      <Card>
        <table className="min-w-full">
          <thead>
            <tr className="text-left text-sm text-gray-500">
              <th className="p-2">名称</th>
              <th className="p-2">来源</th>
              <th className="p-2">版本</th>
              <th className="p-2">最近同步</th>
              <th className="p-2">状态</th>
              <th className="p-2">标签</th>
              <th className="p-2">操作</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr><td className="p-6 text-gray-500" colSpan={7}>暂无数据</td></tr>
            ) : filtered.map((m, idx) => (
              <tr key={`${m.namespace}/${m.name}`} className={idx % 2 === 0 ? "bg-white" : "bg-muted"}>
                <td className="p-2">
                  <Link to={`/models/${m.namespace}/${m.name}`} className="text-primary-700 hover:underline">{m.name}</Link>
                </td>
                <td className="p-2">{m.sourceRef}</td>
                <td className="p-2">{m.versionsReady}/{m.versionsTotal}</td>
                <td className="p-2">{new Date(m.lastSyncTime).toLocaleString()}</td>
                <td className="p-2"><Badge phase={m.status} /></td>
                <td className="p-2">
                  {m.tags?.map((t) => (
                    <span key={t} className="inline-block px-2 py-1 mr-1 rounded-lg bg-gray-100 text-gray-700 text-xs">{t}</span>
                  ))}
                </td>
              <td className="p-2">
                  <div className="flex gap-2">
                    <Button size="sm" variant="outline">查看</Button>
                    <Button size="sm" variant="danger" onClick={() => onDelete(m.name)}>删除</Button>
                  </div>
              </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
      <ConfirmDialog open={!!pendingDelete} title={`删除模型 ${pendingDelete || ""}`} description="此操作将移除模型及其详情（Mock）。" onConfirm={confirmDelete} onCancel={cancelDelete} />
    </div>
  );
}
