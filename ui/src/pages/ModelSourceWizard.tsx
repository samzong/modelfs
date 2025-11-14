import { useEffect, useState } from "react";
import Card from "@components/Card";
import Button from "@components/Button";
import SectionHeader from "@components/SectionHeader";
import { useUiState } from "@app/state";
import { useDataStore } from "@app/dataStore";
import { client } from "@api/client";

const types = ["HUGGING_FACE", "GIT", "HTTP", "S3", "NFS", "PVC"];

export default function ModelSourceWizardPage() {
  const ns = useUiState((s) => s.namespace);
  const setNs = useUiState((s) => s.setNamespace);
  const { namespaces, refreshNamespaces } = useDataStore();
  useEffect(() => { refreshNamespaces(); }, []);
  const path = window.location.pathname;
  const isEdit = path.includes("/edit");
  const parts = path.split("/");
  const editNs = isEdit && parts.length >= 5 ? parts[2] : ns;
  const editName = isEdit && parts.length >= 5 ? parts[3] : "";
  const [name, setName] = useState("");
  const [type, setType] = useState(types[0]);
  const [secretRef, setSecretRef] = useState("");
  const [secretStatus, setSecretStatus] = useState<{ ready: boolean; message: string } | null>(null);
  const [configText, setConfigText] = useState("");
  const [err, setErr] = useState("");
  const config = parseConfig(configText);
  useEffect(() => {
    if (isEdit) {
      client.getModelSource(editNs, editName).then((d) => {
        setName(d.name);
        setNs(d.namespace);
        setType(d.spec?.type || types[0]);
        setSecretRef(d.spec?.secretRef || "");
        setConfigText(Object.entries(d.spec?.config || {}).map(([k,v]) => `${k}=${v}`).join("\n"));
      }).catch(() => {});
    }
  }, [isEdit, editNs, editName]);
  useEffect(() => {
    if (!secretRef) { setSecretStatus(null); return; }
    client.validateSecret(ns, secretRef).then(setSecretStatus).catch(() => setSecretStatus({ ready: false, message: "校验失败" }));
  }, [ns, secretRef]);
  const canSave = name.trim() !== "" && type.trim() !== "";
  async function save() {
    if (!canSave) { setErr("请填写名称与类型"); return; }
    try {
      if (isEdit) {
        await client.updateModelSource(ns, name, { type, secretRef, config });
      } else {
        await client.createModelSource(ns, { name, type, secretRef, config });
      }
      window.location.href = "/modelsources";
    } catch (e) {
      setErr(String(e));
    }
  }
  return (
    <div className="space-y-4">
      <SectionHeader title={isEdit ? "编辑 ModelSource" : "创建 ModelSource"} />
      <div className="flex gap-4">
        <Card className="flex-1">
          <div className="space-y-3">
            <div className="flex gap-3">
              <div className="flex-1">
                <label htmlFor="ms-name" className="block text-sm text-gray-600 mb-1">名称</label>
                <input id="ms-name" className={`form-input w-full ${isEdit ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : ''}`} value={name} onChange={(e) => setName(e.target.value)} disabled={isEdit} />
              </div>
              <div className="flex-1">
                <label htmlFor="ms-ns" className="block text-sm text-gray-600 mb-1">命名空间</label>
                <select id="ms-ns" className={`form-select w-full ${isEdit ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : ''}`} value={ns} onChange={(e) => setNs(e.target.value)} disabled={isEdit}>
                  {namespaces.map((n) => (<option key={n.name} value={n.name}>{n.name}</option>))}
                </select>
              </div>
            </div>
            <div className="flex gap-3">
              <div className="flex-1">
                <label htmlFor="ms-type" className="block text-sm text-gray-600 mb-1">类型</label>
                <select id="ms-type" className={`form-select w-full ${isEdit ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : ''}`} value={type} onChange={(e) => setType(e.target.value)} disabled={isEdit}>
                  {types.map((t) => (<option key={t} value={t}>{t}</option>))}
                </select>
              </div>
              <div className="flex-1">
                <label htmlFor="ms-secret" className="block text-sm text-gray-600 mb-1">SecretRef</label>
                <input id="ms-secret" className="form-input w-full" value={secretRef} onChange={(e) => setSecretRef(e.target.value)} />
                {secretStatus ? (
                  <div className="text-xs mt-1">{secretStatus.ready ? <span className="text-green-700">就绪</span> : <span className="text-red-700">{secretStatus.message || "不可用"}</span>}</div>
                ) : null}
              </div>
            </div>
            <div>
              <label htmlFor="ms-config" className="block text-sm text-gray-600 mb-1">配置（key=value，每行一项）</label>
              <textarea id="ms-config" className="form-input w-full" rows={6} value={configText} onChange={(e) => setConfigText(e.target.value)} />
            </div>
            {err ? <div className="text-red-600 text-sm">{err}</div> : null}
          </div>
        </Card>
        <Card className="w-64 h-fit">
          <div className="space-y-2">
            <Button onClick={save} disabled={!canSave}>保存</Button>
          </div>
        </Card>
      </div>
    </div>
  );
}

function parseConfig(text: string): Record<string, string> {
  const out: Record<string, string> = {};
  text.split(/\n+/).map((l) => l.trim()).filter(Boolean).forEach((line) => {
    const idx = line.indexOf("=");
    if (idx > 0) {
      const k = line.slice(0, idx).trim();
      const v = line.slice(idx + 1).trim();
      if (k) out[k] = v;
    }
  });
  return out;
}
