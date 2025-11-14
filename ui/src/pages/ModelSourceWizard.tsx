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
  const [name, setName] = useState("");
  const [type, setType] = useState(types[0]);
  const [secretRef, setSecretRef] = useState("");
  const [configText, setConfigText] = useState("");
  const [err, setErr] = useState("");
  const config = parseConfig(configText);
  const canSave = name.trim() !== "" && type.trim() !== "";
  async function save() {
    if (!canSave) { setErr("请填写名称与类型"); return; }
    try {
      await client.createModelSource(ns, { name, type, secretRef, config });
      window.location.href = "/modelsources";
    } catch (e) {
      setErr(String(e));
    }
  }
  return (
    <div className="space-y-4">
      <SectionHeader title="创建 ModelSource" />
      <div className="flex gap-4">
        <Card className="flex-1">
          <div className="space-y-3">
            <div className="flex gap-3">
              <div className="flex-1">
                <label htmlFor="ms-name" className="block text-sm text-gray-600 mb-1">名称</label>
                <input id="ms-name" className="form-input w-full" value={name} onChange={(e) => setName(e.target.value)} />
              </div>
              <div className="flex-1">
                <label htmlFor="ms-ns" className="block text-sm text-gray-600 mb-1">命名空间</label>
                <select id="ms-ns" className="form-select w-full" value={ns} onChange={(e) => setNs(e.target.value)}>
                  {namespaces.map((n) => (<option key={n.name} value={n.name}>{n.name}</option>))}
                </select>
              </div>
            </div>
            <div className="flex gap-3">
              <div className="flex-1">
                <label htmlFor="ms-type" className="block text-sm text-gray-600 mb-1">类型</label>
                <select id="ms-type" className="form-select w-full" value={type} onChange={(e) => setType(e.target.value)}>
                  {types.map((t) => (<option key={t} value={t}>{t}</option>))}
                </select>
              </div>
              <div className="flex-1">
                <label htmlFor="ms-secret" className="block text-sm text-gray-600 mb-1">SecretRef</label>
                <input id="ms-secret" className="form-input w-full" value={secretRef} onChange={(e) => setSecretRef(e.target.value)} />
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

