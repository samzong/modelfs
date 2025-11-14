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
      <SectionHeader title="Models" description="Models in Namespace" right={<Link to="/models/wizard"><Button variant="primary">Create</Button></Link>} />
      <div className="toolbar">
        <input className="form-input w-72" placeholder="Filter by name or tag" value={filterText} onChange={(e) => setFilterText(e.target.value)} />
      </div>
      <ErrorBanner items={errors} />
      <Card>
        <table className="min-w-full">
          <thead>
            <tr className="text-left text-sm text-gray-500">
              <th className="p-2">Name</th>
              <th className="p-2">Source</th>
              <th className="p-2">Version</th>
              <th className="p-2">Last Sync</th>
              <th className="p-2">Status</th>
              <th className="p-2">Tags</th>
              <th className="p-2">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr><td className="p-6 text-gray-500" colSpan={7}>No Data</td></tr>
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
                    <Link to={`/models/${m.namespace}/${m.name}`}>
                      <Button size="sm" variant="outline">View</Button>
                    </Link>
                    <Link to={`/models/${m.namespace}/${m.name}/edit`}>
                      <Button size="sm" variant="secondary">Edit</Button>
                    </Link>
                    <Button size="sm" variant="danger" onClick={() => onDelete(m.name)}>Delete</Button>
                  </div>
              </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
      <ConfirmDialog open={!!pendingDelete} title={`Delete Model ${pendingDelete || ""}`} description="This operation removes the model and its details (Mock)." onConfirm={confirmDelete} onCancel={cancelDelete} />
    </div>
  );
}
