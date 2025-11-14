import { useEffect, useState } from "react";
import Card from "@components/Card";
import { client } from "@api/client";

export default function ModelSourceDetailPage() {
  const parts = window.location.pathname.split("/");
  const ns = parts[2];
  const name = parts[3];
  const [detail, setDetail] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");
  useEffect(() => {
    setLoading(true);
    client.getModelSource(ns, name)
      .then((d) => { setDetail(d); setErr(""); })
      .catch((e) => { setErr(String(e)); setDetail(null); })
      .finally(() => setLoading(false));
  }, [ns, name]);
  if (loading) return <div className="card p-4">加载中...</div>;
  if (err) return <div className="card p-4 text-red-600">{err}</div>;
  if (!detail) return <div className="card p-4">无数据</div>;
  return (
    <div className="space-y-4">
      <Card>
        <div className="flex items-start justify-between">
          <div>
            <h2 className="text-2xl font-semibold">{detail.name}</h2>
            <p className="text-gray-600">命名空间：{detail.namespace}</p>
          </div>
          <a href={`/modelsources/${detail.namespace}/${detail.name}/edit`} className="px-3 py-1 rounded-lg border">编辑</a>
        </div>
      </Card>
      <Card>
        <h3 className="text-lg font-semibold mb-2">配置</h3>
        <pre className="bg-muted p-3 rounded-lg text-sm overflow-auto">{JSON.stringify(detail.spec, null, 2)}</pre>
      </Card>
    </div>
  );
}
