# modelfs KEP：基于 BaizeAI/dataset 的极简模型权重管理

## 目标
1. 最小 CRD 集：`ModelSource`（凭证/连接）与 `Model`（模型+版本+同步期望）。ModelSource 统一管理访问凭证与公共配置；Model 负责声明每个版本的目标状态。
2. Model 以 **declarative** 方式描述版本所需的 Dataset（state、PVC 规格、共享策略等），modelfs Controller 负责创建/更新/删除 BaizeAI `Dataset` 与其 REFERENCE 副本。
3. 保障安全与可观测性：凭证复用、跨命名空间共享的 RBAC 控制、完整的版本状态（phase/conditions/pvc 信息）以及明确的生命周期策略。

## 背景与动机
- BaizeAI/dataset 擅长下载与 PVC 生命周期，但缺少“模型目录、版本元信息、共享策略”以及凭证复用能力。
- 先前 4 个 CRD（ModelSource/Model/ModelSync/ModelReference）导致 imperative 工作流与重复概念。评审建议将同步意图内联到 Model 中，保持 API 简洁。
- 本 KEP 通过单一 Model CRD 即可表达“我希望版本 X 在集群内可用/共享/占用多少存储”，其余工作由 controller 与 BaizeAI/dataset 完成。

## 设计
### 1. ModelSource
```yaml
apiVersion: model.samzong.dev/v1
kind: ModelSource
spec:
  type: HUGGING_FACE
  secretRef: hf-token
  config:
    endpoint: https://huggingface.co
    include: "*.safetensors"
```
- `secretRef` 必填，controller 仅具有“同 namespace Secret 只读”权限；Secret 缺失或格式错误会在 `status.conditions` 中暴露 `CredentialsReady=false`。
- `config` 仅允许对应 DatasetType 支持的键（Admission webhook 校验）。
- ModelSource 不由 Model 拥有。删除流程：ModelSource finalizer 使用 field index(`modelSourceRef`) 查找所有引用的 Model；当引用清零时才允许删除。

### 2. Model（合并同步与共享语义）
```yaml
apiVersion: model.samzong.dev/v1
kind: Model
metadata:
  name: qwen3
spec:
  sourceRef: huggingface-source
  display:
    description: "Qwen3 家族"
    tags: ["qwen","chat"]
    logoURL: https://...
  versions:
    - name: fp16
      repo: qwen/Qwen3-7B
      revision: main
      precision: FP16
      metadata:
        tags: ["fp16","full"]
      storage:
        accessModes: ["ReadWriteMany"]
        resources:
          requests:
            storage: 12Ti
        storageClassName: fast-rwx
      state: PRESENT        # 默认 PRESENT，可显式设为 ABSENT
      share:
        enabled: true
        namespaceSelector:
          matchLabels:
            tenant: inference
        requireOptInLabel: modelfs.samzong.dev/share-opt-in
    - name: q4
      repo: qwen/Qwen3-7B
      precision: INT4
      state: ABSENT
```
- `versions` 为 slice；Admission webhook：非空、`name` 唯一、`state` 默认 `PRESENT`。
- `storage` 使用精简类型 `ModelVolumeSpec`（仅包含 accessModes / resources / storageClassName），避免暴露 PVC 其他敏感字段。
- 每个版本可拥有独立 `metadata`，用来覆盖/补充 `spec.display.tags` 等顶层信息，满足 catalog 需求。
- `share.requireOptInLabel` 指定命名空间必须携带 `key=value` 才能接收共享；这样 selector + opt-in 双重保护。

### 3. 控制器行为
#### 3.1 主 Dataset 生命周期
1. Reconcile 触发源：Model、ModelSource、Dataset 以及 Namespace（用于 share selector）。
2. 对于 `state=PRESENT` 的版本：
   - 根据 `{namespace}/{modelName}/{versionName}` 计算唯一 Dataset 名称 `mdl-<model>-<version>`。
   - 调用 `pkg/dataset.BuildDatasetSpec` 构造 spec：使用 ModelVersion.repo/revision、ModelSource config + Secret options、ModelVolumeSpec；若 Secret 不可用则设置版本 phase=Failed 并重试。
   - `EnsureDataset`：若 Dataset 不存在则创建，存在则 patch（元数据更新、PVC 扩容、共享字段等）。
3. 对于 `state=ABSENT`：如果 Dataset 存在，先标记 ModelVersion status `observedState=ABSENT`，再删除 Dataset。等待 BaizeAI controller 清理 PVC 后，版本 Ready=false。
4. Model 对 Dataset 使用 ownerReference（同 namespace）+ finalizer `modelfs.samzong.dev/model-finalizer`。删除 Model 时：先删除所有 reference Dataset，再删主 Dataset，最后移除 finalizer；若资源清理超时（默认 1h）则在 status 中暴露 blocking condition，允许管理员强制释放。

#### 3.2 共享与 Namespace 监听
- 当 `share.enabled=true`：
  1. 主 Dataset `spec.share=true` 并写入 selector。
  2. Controller watch Namespace 资源，筛选满足 selector 且带 opt-in label 的 namespace；对每个 namespace 创建/更新 `Dataset`：
     ```yaml
     metadata:
       name: share-<source-ns>-<model>-<version>
       labels:
         modelfs.samzong.dev/model: <source-ns>/<model>
         modelfs.samzong.dev/version: <version>
     spec:
       source:
         type: REFERENCE
         uri: dataset://<source-ns>/<mdl-...>
     ```
  3. **跨命名空间 OwnerReference 不允许**：这些 REFERENCE Dataset 通过标签与 Model 关联，Model finalizer 记录它们的 namespace/名称；删除 Model 时主动遍历删除。
  4. 如果目标 namespace 已存在同名非 modelfs Dataset，写入 Model status 的该版本 `conditions`（type=`ShareConflict`）并停止下发，避免覆写用户资源。

#### 3.3 版本更新策略
- **元数据变更（description/tags）**：只 patch Model status / Dataset annotation，无需重建 PVC。
- **PVC 容量增加**：优先尝试对现有 PVC 执行在线扩容（patch `resources.requests.storage`），若 CSI 不支持或缩容需求出现，则执行蓝绿策略。
- **repo/revision/Source 变更**：
  1. 创建新 Dataset `mdl-<model>-<version>-<hash>`（hash 基于 repo+revision）。
  2. 等待新 Dataset READY，更新 Model status 中的 `activeDataset` 指向新 PVC。
  3. 标记旧 Dataset `phase=Retiring`，在 configurable 的保留期后删除，确保零停机。

#### 3.4 状态同步
- `Model.status.syncedVersions`（按 version name 匹配）包含：phase、pvcName、activeDataset、lastSyncTime、全部 Dataset conditions、observedState、observedStorage（resource.Quantity）。
- 对于已从 spec 删除的版本，status 条目保留但标记 `observedState=ABSENT`，直到残留资源清理完毕。
- `status.observedGeneration` 跟踪 Model 级 generation；版本级 condition 通过 `observedVersionHash` 字段描述最新观测的 spec 哈希。

### 4. 与 BaizeAI/dataset 的互操作
- `pkg/dataset` 组件：
  - `BuildDatasetSpec(ctx, client, version, source) (*datasetv1alpha1.DatasetSpec, error)`：读取 Secret -> options、repo -> URI（HUGGING_FACE/HTTP/S3 等），返回 Dataset spec。
  - `EnsureDataset`/`EnsureReferenceDataset`：封装 create-or-patch，并附带遥测；partial failure 会返回 structured error，controller 会写入 Model status。
- 继续依赖 dataset controller 的 PVC/job/finalizer 逻辑；modelfs 不重新实现下载流程。

### 5. 安全与生命周期细节
- **Secrets**：Admission webhook 验证 Secret 格式；Controller 访问日志记录 Secret 名称（不含值）。Secret 丢失会导致版本 phase=Failed 并重试。
- **share selector 权限**：更改 `share.namespaceSelector` 或 opt-in label 需要具有 `modelfs.samzong.dev/share-admin` ClusterRole；webhook 校验并审计。
- **Namespace opt-in**：默认 label `modelfs.samzong.dev/share-opt-in: "true"`；仅集群管理员可设置，避免租户随意加入。
- **ModelSource 删除**：webhook 阻止删除仍被引用的 source；finalizer 通过 field index实时感知引用数量。

### 6. 观测与调试
- Metrics：
  - `modelfs_modelversion_phase{phase}`
  - `modelfs_dataset_reconcile_seconds_bucket`
  - `modelfs_share_target_total`
  - `modelfs_status_conflicts_total`（share 冲突、Secret 错误等）
- `kubectl get model <name> -o yaml` 即可查看版本级状态、activeDataset、share 冲突；无需直接查看 Dataset。

## 兼容与扩展
- 可追加 `ModelVersion.triggerPolicy` 或 `state=PAUSED` 等模式；当前语义保持向后兼容。
- 如果未来需要显式的共享请求对象，可新增 `ModelReferenceRequest` CRD，实现与现有 Model API 相互独立。
- 模板/patch 机制、自动滚动升级策略可在后续 KEP 中引入而不破坏现有字段。

## 风险与缓解
| 风险 | 缓解 |
| --- | --- |
| PVC 规格配置错误或无默认值 | Admission webhook 强制 `storage.requests` 有效、可设置全局默认 storageClass；状态中展示实际容量。 |
| Secret 丢失/权限不足 | ModelSource status 条件提示，Model version phase=Failed；controller 指数退避重试并输出结构化错误日志。 |
| 跨 namespace 滥用 | share selector 修改需 RBAC；Namespace 必须具备 opt-in label；Model finalizer 手动删除所有 reference Dataset。 |
| 状态冲突/patch 失败 | 所有 status 更新使用 server-side patch + exponential backoff（初始100ms，最大5s，重试10次）。 |
| Dataset 变更导致停机 | 针对 repo/revision 采用蓝绿策略；PVC 扩容优先在线 patch，无法扩容时才新建。 |
| ModelSource 删除时存在引用 | field index + finalizer，status.referencedBy 列出引用者，管理员可审计。 |

## 后续工作
1. 实现 Admission Webhook：验证 `ModelSource.config`、`versions` 唯一性、`state` 默认、`ModelVolumeSpec` 范围以及 share selector 权限。
2. e2e 测试：覆盖 present/absent 切换、PVC 扩容、蓝绿替换、share 冲突与删除流程。
3. 提供 CLI/可视化工具，展示 ModelVersion 状态与跨 namespace share 映射，辅助 SRE 诊断。
