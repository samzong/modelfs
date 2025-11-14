import { useEffect, useMemo, useState } from "react";
import * as Select from "@radix-ui/react-select";
import Card from "@components/Card";
import Button from "@components/Button";
import SectionHeader from "@components/SectionHeader";
import { models, modelDetails } from "@mocks/data";
import { useDataStore } from "@app/dataStore";
import { useUiState } from "@app/state";

type VersionInput = {
  name: string;
  repo: string;
  revision?: string;
  precision?: "FP16" | "INT4" | "INT8";
  desiredState: "PRESENT" | "ABSENT";
  shareEnabled: boolean;
};

export default function ModelWizardPage() {
  const path = window.location.pathname;
  const isEdit = path.includes("/edit");
  const parts = path.split("/");
  const maybeNs = parts.length >= 5 ? parts[2] : "model-system";
  const maybeName = parts.length >= 5 ? parts[3] : "";

  const [step, setStep] = useState<number>(1);
  const [name, setName] = useState<string>(isEdit ? maybeName : "");
  const ns = useUiState((s) => s.namespace);
  const setNs = useUiState((s) => s.setNamespace);
  const [sourceRef, setSourceRef] = useState<string>("");
  const [tags, setTags] = useState<string[]>([]);
  const [description, setDescription] = useState<string>("");
  const [versions, setVersions] = useState<VersionInput[]>([{ name: "", repo: "", desiredState: "PRESENT", shareEnabled: false }]);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { namespaces, sources, refreshNamespaces, refreshAll } = useDataStore();
  useEffect(() => { refreshNamespaces(); }, []);
  useEffect(() => { refreshAll(ns); }, [ns]);
  useEffect(() => {
    const list = sources.filter((s) => s.namespace === ns);
    if (!sourceRef && list.length) {
      setSourceRef(list[0].name);
    }
  }, [sources, ns]);

  const canNext1 = name.trim() !== "" && sourceRef.trim() !== "";
  const canNext2 = versions.length > 0 && versions.every(v => v.name.trim() !== "" && v.repo.trim() !== "");

  function addVersion() {
    setVersions(prev => [...prev, { name: "", repo: "", desiredState: "PRESENT", shareEnabled: false }]);
  }
  function updateVersion(i: number, patch: Partial<VersionInput>) {
    setVersions(prev => prev.map((v, idx) => (idx === i ? { ...v, ...patch } : v)));
  }
  function removeVersion(i: number) {
    setVersions(prev => prev.filter((_, idx) => idx !== i));
  }

  function next() {
    if (step === 1 && !canNext1) { setErrors({ basic: "请填写名称与来源" }); return; }
    if (step === 2 && !canNext2) { setErrors({ versions: "每个版本需填写名称与仓库" }); return; }
    setErrors({});
    setStep(step + 1);
  }
  function prev() { setStep(step - 1); }

  function save() {
    const summary = {
      name,
      namespace: ns,
      sourceRef,
      tags,
      versionsReady: 0,
      versionsTotal: versions.length,
      lastSyncTime: new Date().toISOString(),
      status: "PENDING" as const,
    };
    const detail = {
      summary,
      description,
      versions: versions.map(v => ({
        name: v.name,
        repo: v.repo,
        desiredState: v.desiredState,
        shareEnabled: v.shareEnabled,
        datasetPhase: "PENDING" as const,
      }))
    };
    const key = `${ns}/${name}`;
    const idx = models.findIndex(m => m.namespace === ns && m.name === name);
    if (idx >= 0) {
      models[idx] = summary as any;
      (modelDetails as any)[key] = detail as any;
    } else {
      models.push(summary as any);
      (modelDetails as any)[key] = detail as any;
    }
    window.location.href = "/models";
  }

  const sourcesFiltered = useMemo(() => sources.filter((s) => s.namespace === ns), [sources, ns]);
  return (
    <div className="space-y-4">
      <SectionHeader title={isEdit ? "编辑模型" : "创建模型"} description="三步完成模型创建或编辑" />
      <div className="flex gap-4">
        <Card className="flex-1">
          {step === 1 && (
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="flex-1">
                  <label htmlFor="model-name" className="block text-sm text-gray-600 mb-1">名称</label>
                  <input id="model-name" className="form-input w-full" value={name} onChange={(e) => setName(e.target.value)} />
                </div>
                <div className="flex-1">
                  <label htmlFor="model-ns" className="block text-sm text-gray-600 mb-1">命名空间</label>
                  <Select.Root value={ns} onValueChange={setNs}>
                    <Select.Trigger id="model-ns" className="border px-3 h-10 rounded-lg w-full text-left">
                      <Select.Value placeholder={ns} />
                    </Select.Trigger>
                    <Select.Content className="bg-white border rounded-lg shadow z-50">
                      <Select.Viewport>
                        {namespaces.map((n) => (
                          <Select.Item key={n.name} value={n.name} className="px-3 py-2">
                            <Select.ItemText>{n.name}</Select.ItemText>
                          </Select.Item>
                        ))}
                      </Select.Viewport>
                    </Select.Content>
                  </Select.Root>
                </div>
              </div>
              <div>
                <label htmlFor="model-src" className="block text-sm text-gray-600 mb-1">来源（SourceRef）</label>
                <Select.Root value={sourceRef} onValueChange={setSourceRef}>
                  <Select.Trigger id="model-src" className="border px-3 h-10 rounded-lg w-full text-left">
                    <Select.Value placeholder={sourceRef || "请选择来源"} />
                  </Select.Trigger>
                  <Select.Content className="bg-white border rounded-lg shadow z-50">
                    <Select.Viewport>
                      {sourcesFiltered.map((ms) => (
                        <Select.Item key={ms.name} value={ms.name} className="px-3 py-2">
                          <Select.ItemText>{ms.name}</Select.ItemText>
                        </Select.Item>
                      ))}
                    </Select.Viewport>
                  </Select.Content>
                </Select.Root>
              </div>
              <div>
                <label htmlFor="model-desc" className="block text-sm text-gray-600 mb-1">描述</label>
                <textarea id="model-desc" className="form-input w-full" value={description} onChange={(e) => setDescription(e.target.value)} />
              </div>
              <div>
                <label htmlFor="model-tags" className="block text-sm text-gray-600 mb-1">标签（逗号分隔）</label>
                <input id="model-tags" className="form-input w-full" value={tags.join(',')} onChange={(e) => setTags(e.target.value.split(',').map(s => s.trim()).filter(Boolean))} />
              </div>
              {errors.basic ? <div className="text-red-600 text-sm">{errors.basic}</div> : null}
            </div>
          )}
          {step === 2 && (
            <div className="space-y-3">
              {versions.map((v, i) => (
                <div key={i} className="border rounded-lg p-3">
                  <div className="flex gap-3">
                    <input className="form-input flex-1" placeholder="版本名" value={v.name} onChange={(e) => updateVersion(i, { name: e.target.value })} />
                    <input className="form-input flex-1" placeholder="仓库" value={v.repo} onChange={(e) => updateVersion(i, { repo: e.target.value })} />
                  </div>
                  <div className="flex gap-3 mt-2">
                    <select className="form-select" value={v.precision || ""} onChange={(e) => updateVersion(i, { precision: (e.target.value || undefined) as any })}>
                      <option value="">Precision</option>
                      <option value="FP16">FP16</option>
                      <option value="INT4">INT4</option>
                      <option value="INT8">INT8</option>
                    </select>
                    <select className="form-select" value={v.desiredState} onChange={(e) => updateVersion(i, { desiredState: e.target.value as any })}>
                      <option value="PRESENT">PRESENT</option>
                      <option value="ABSENT">ABSENT</option>
                    </select>
                    <label className="inline-flex items-center gap-2">
                      <input type="checkbox" checked={v.shareEnabled} onChange={(e) => updateVersion(i, { shareEnabled: e.target.checked })} /> 分享
                    </label>
                    <Button variant="outline" size="sm" onClick={() => removeVersion(i)}>移除</Button>
                  </div>
                </div>
              ))}
              <Button variant="secondary" onClick={addVersion}>添加版本</Button>
              {errors.versions ? <div className="text-red-600 text-sm">{errors.versions}</div> : null}
            </div>
          )}
          {step === 3 && (
            <div className="space-y-3">
              <div className="text-sm text-gray-600">预览</div>
              <pre className="bg-muted p-3 rounded-lg text-sm overflow-auto">{JSON.stringify({
                apiVersion: "model.samzong.dev/v1",
                kind: "Model",
                metadata: { name, namespace: ns, labels: tags.reduce((acc, t) => ({ ...acc, [t]: "true" }), {}) },
                spec: { sourceRef, versions },
              }, null, 2)}</pre>
            </div>
          )}
        </Card>
        <Card className="w-64 h-fit">
          <div className="space-y-2">
            <div>步骤</div>
            <div className="flex flex-col gap-2">
              <Button variant={step === 1 ? "primary" : "outline"} onClick={() => setStep(1)}>1 基础</Button>
              <Button variant={step === 2 ? "primary" : "outline"} onClick={() => setStep(2)}>2 版本</Button>
              <Button variant={step === 3 ? "primary" : "outline"} onClick={() => setStep(3)}>3 预览</Button>
            </div>
            <div className="flex gap-2 pt-2">
              {step > 1 ? <Button variant="outline" onClick={prev}>上一步</Button> : null}
              {step < 3 ? <Button onClick={next} disabled={(step === 1 && !canNext1) || (step === 2 && !canNext2)}>下一步</Button> : <Button onClick={save} disabled={!canNext1 || !canNext2}>保存</Button>}
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
