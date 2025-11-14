import { useUiState } from "@app/state";
import Card from "@components/Card";
import Badge from "@components/Badge";
import SectionHeader from "@components/SectionHeader";
import Button from "@components/Button";
import { useEffect, useMemo, useState } from "react";
import ConfirmDialog from "@components/ConfirmDialog";
import { useDataStore } from "@app/dataStore";
import { Link } from "@tanstack/react-router";

export default function ModelSourcesPage() {
  const ns = useUiState((s) => s.namespace);
  const [refresh, setRefresh] = useState(0);
  const { sources, refreshAll } = useDataStore();
  useEffect(() => { refreshAll(ns); }, [ns]);
  const rows = useMemo(() => sources.filter((s) => s.namespace === ns), [sources, ns, refresh]);
  const [q, setQ] = useState("");
  const filtered = useMemo(() => rows.filter((s) => !q || s.name.includes(q) || s.type.includes(q) || (s.secretRef || "").includes(q)), [rows, q]);
  const [pendingDelete, setPendingDelete] = useState<string | null>(null);
  function onDelete(name: string) { setPendingDelete(name); }
  function confirmDelete() {
    if (!pendingDelete) return;
    const idx = rows.findIndex(s => s.namespace === ns && s.name === pendingDelete);
    if (idx >= 0 && !(rows[idx].referencedModels && rows[idx].referencedModels!.length > 0)) {
      setRefresh(refresh + 1);
    }
    setPendingDelete(null);
  }
  function cancelDelete() { setPendingDelete(null); }
  return (
    <div className="space-y-4">
      <SectionHeader title="ModelSources" description="Model Source Configuration and Credential Status" right={<a href="/modelsources/new"><Button variant="primary">Create</Button></a>} />
      <div className="toolbar">
        <input className="form-input w-72" placeholder="Search name/type/Secret" value={q} onChange={(e) => setQ(e.target.value)} />
      </div>
      <Card>
        <table className="min-w-full">
          <thead>
            <tr className="text-left text-sm text-gray-500">
              <th className="p-2">Name</th>
              <th className="p-2">Type</th>
              <th className="p-2">Secret</th>
              <th className="p-2">Credentials</th>
              <th className="p-2">Referenced Models</th>
              <th className="p-2">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr><td className="p-6 text-gray-500" colSpan={5}>No Sources</td></tr>
            ) : filtered.map((s, idx) => (
              <tr key={`${s.namespace}/${s.name}`} className={idx % 2 === 0 ? "bg-white" : "bg-muted"}>
                <td className="p-2">
                  <Link to={`/modelsources/${s.namespace}/${s.name}`} className="text-primary-700 hover:underline">{s.name}</Link>
                </td>
                <td className="p-2">{s.type}</td>
                <td className="p-2">{s.secretRef || "-"}</td>
                <td className="p-2">{s.credentialsReady ? <span className="px-2 py-1 rounded-lg bg-green-100 text-green-800 text-sm">Ready</span> : s.credentialsStatus || "Unknown"}</td>
                <td className="p-2">{s.referencedModels?.length ? s.referencedModels?.join(", ") : "-"}</td>
                <td className="p-2">
                  <div className="flex gap-2">
                    <Link to={`/modelsources/${s.namespace}/${s.name}`}>
                      <Button size="sm" variant="outline">View</Button>
                    </Link>
                    <Link to={`/modelsources/${s.namespace}/${s.name}/edit`}>
                      <Button size="sm" variant="secondary">Edit</Button>
                    </Link>
                    <Button size="sm" variant="danger" onClick={() => onDelete(s.name)} disabled={!!(s.referencedModels && s.referencedModels.length > 0)}>Delete</Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
      <ConfirmDialog open={!!pendingDelete} title={`Delete ModelSource ${pendingDelete || ""}`} description="Cannot delete when referenced (Mock)." onConfirm={confirmDelete} onCancel={cancelDelete} />
    </div>
  );
}
