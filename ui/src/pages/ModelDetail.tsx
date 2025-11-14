import { useEffect, useMemo, useState } from "react";
import { client } from "@api/client";
import { useUiState } from "@app/state";
import Card from "@components/Card";
import Badge from "@components/Badge";

export default function ModelDetailPage() {
  const ns = useUiState((s) => s.namespace);
  const name = window.location.pathname.split("/").slice(-1)[0];
  const [detail, setDetail] = useState<any>(null);
  useEffect(() => { client.getModel(ns, name).then(setDetail).catch(() => setDetail(null)); }, [ns, name]);
  const [expanded, setExpanded] = useState<string | null>(null);
  const [tab, setTab] = useState<"info"|"yaml">("info");
  const yaml = useMemo(() => (detail ? toYAML(detail) : ""), [detail]);
  if (!detail) return <div className="card p-4">Loading...</div>;
  return (
    <div className="space-y-4">
      <Card>
        <div className="flex items-start justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{detail.summary.name}</h2>
            {detail.description ? <p className="text-gray-600">{detail.description}</p> : null}
          </div>
          <div className="flex items-center gap-2">
            <Badge phase={detail.summary.status} />
            <a href={`/models/${detail.summary.namespace}/${detail.summary.name}/edit`} className="px-3 py-1 rounded-lg border">Edit</a>
          </div>
        </div>
        {detail.summary.tags?.length ? (
          <div className="mt-2">
            {detail.summary.tags?.map((t) => (
              <span key={t} className="inline-block px-2 py-1 mr-1 rounded-lg bg-gray-100 text-gray-700 text-xs">{t}</span>
            ))}
          </div>
        ) : null}
      </Card>
      <Card>
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-lg font-semibold">Versions</h3>
          <div className="flex gap-2">
            <button className={`px-3 py-1 rounded ${tab === "info" ? "bg-primary-600 text-white" : "bg-muted"}`} onClick={() => setTab("info")}>Info</button>
            <button className={`px-3 py-1 rounded ${tab === "yaml" ? "bg-primary-600 text-white" : "bg-muted"}`} onClick={() => setTab("yaml")}>YAML</button>
          </div>
        </div>
        {tab === "yaml" ? (
          <pre className="bg-muted p-3 rounded-lg text-sm overflow-auto">{yaml}</pre>
        ) : (
        <div>
        <table className="min-w-full">
          <thead>
            <tr className="text-left text-sm text-gray-500">
              <th className="p-2">Version</th>
              <th className="p-2">Repo</th>
              <th className="p-2">Expectation</th>
              <th className="p-2">Share</th>
              <th className="p-2">Dataset</th>
              <th className="p-2">PVC</th>
              <th className="p-2">Storage (Observability)</th>
            </tr>
          </thead>
          <tbody>
            {detail.versions.map((v, idx) => (
              <tr key={v.name} className={idx % 2 === 0 ? "bg-white" : "bg-muted"} onClick={() => setExpanded(expanded === v.name ? null : v.name)}>
                <td className="p-2">{v.name}</td>
                <td className="p-2">{v.repo}</td>
                <td className="p-2">{v.desiredState}</td>
                <td className="p-2">{v.shareEnabled ? "Enabled" : "Disabled"}</td>
                <td className="p-2">{v.datasetPhase}</td>
                <td className="p-2">{v.pvcName || "-"}</td>
                <td className="p-2">{v.observedStorage || "-"}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {expanded ? (
          <div className="mt-3 border rounded-lg p-3">
            <div className="text-sm text-gray-700">Version Details</div>
            <pre className="bg-muted p-2 rounded text-xs overflow-auto">{JSON.stringify(detail.versions.find(v => v.name === expanded), null, 2)}</pre>
            <div className="text-xs text-gray-600 mt-2">kubectl -n {detail.summary.namespace} get pvc {detail.versions.find(v => v.name === expanded)?.pvcName || "-"}</div>
          </div>
        ) : null}
        </div>
        )}
      </Card>
    </div>
  );
}

function toYAML(detail: any): string {
  const lines: string[] = [];
  lines.push("apiVersion: model.samzong.dev/v1");
  lines.push("kind: Model");
  lines.push("metadata:");
  lines.push(`  name: ${detail.summary.name}`);
  lines.push(`  namespace: ${detail.summary.namespace}`);
  if (detail.summary.tags && detail.summary.tags.length) {
    lines.push("  labels:");
    detail.summary.tags.forEach((t: string) => lines.push(`    ${t}: "true"`));
  }
  lines.push("spec:");
  lines.push(`  sourceRef: ${detail.summary.sourceRef}`);
  if (detail.description) {
    lines.push("  display:");
    lines.push(`    description: ${detail.description}`);
  }
  lines.push("  versions:");
  detail.versions.forEach((v: any) => {
    lines.push("  - name: " + v.name);
    lines.push("    repo: " + v.repo);
    if (v.revision) lines.push("    revision: " + v.revision);
    if (v.precision) lines.push("    precision: " + v.precision);
    lines.push("    state: " + v.desiredState);
    lines.push("    share:");
    lines.push("      enabled: " + (v.shareEnabled ? "true" : "false"));
  });
  return lines.join("\n");
}
