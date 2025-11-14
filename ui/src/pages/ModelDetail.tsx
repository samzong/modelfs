import { useMemo, useState } from "react";
import { modelDetails } from "@mocks/data";
import { useUiState } from "@app/state";
import Card from "@components/Card";
import Badge from "@components/Badge";

export default function ModelDetailPage() {
  const ns = useUiState((s) => s.namespace);
  const name = window.location.pathname.split("/").slice(-1)[0];
  const key = `${ns}/${name}`;
  const detail = useMemo(() => modelDetails[key], [key]);
  const [expanded, setExpanded] = useState<string | null>(null);
  if (!detail) return <div>未找到模型</div>;
  return (
    <div className="space-y-4">
      <Card>
        <div className="flex items-start justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{detail.summary.name}</h2>
            {detail.description ? <p className="text-gray-600">{detail.description}</p> : null}
          </div>
          <Badge phase={detail.summary.status} />
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
        <h3 className="text-lg font-semibold mb-2">版本</h3>
        <table className="min-w-full">
          <thead>
            <tr className="text-left text-sm text-gray-500">
              <th className="p-2">版本</th>
              <th className="p-2">Repo</th>
              <th className="p-2">期望</th>
              <th className="p-2">分享</th>
              <th className="p-2">Dataset</th>
              <th className="p-2">PVC</th>
            </tr>
          </thead>
          <tbody>
            {detail.versions.map((v, idx) => (
              <tr key={v.name} className={idx % 2 === 0 ? "bg-white" : "bg-muted"} onClick={() => setExpanded(expanded === v.name ? null : v.name)}>
                <td className="p-2">{v.name}</td>
                <td className="p-2">{v.repo}</td>
                <td className="p-2">{v.desiredState}</td>
                <td className="p-2">{v.shareEnabled ? "启用" : "关闭"}</td>
                <td className="p-2">{v.datasetPhase}</td>
                <td className="p-2">{v.pvcName || "-"}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {expanded ? (
          <div className="mt-3 border rounded-lg p-3">
            <div className="text-sm text-gray-700">版本详情</div>
            <pre className="bg-muted p-2 rounded text-xs overflow-auto">{JSON.stringify(detail.versions.find(v => v.name === expanded), null, 2)}</pre>
            <div className="text-xs text-gray-600 mt-2">kubectl -n {detail.summary.namespace} get pvc {detail.versions.find(v => v.name === expanded)?.pvcName || "-"}</div>
          </div>
        ) : null}
      </Card>
    </div>
  );
}
