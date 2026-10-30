# Runner 与引擎适配契约

状态：草案。用于指导实现，不表示已有代码。

## 1. 目标

让平台不依赖某个扫描器内部模型，同时允许每个引擎保留其特殊能力。适配器分为两层：

1. **Runner 适配器**：调用具体引擎，处理进程、配置、取消、日志；
2. **结果映射器**：把引擎报告转换成统一结果模型。

## 2. 输入

```json
{
  "contract_version": "1.0",
  "scan_id": "uuid",
  "attempt": 1,
  "type": "sca",
  "project_id": "uuid",
  "profile": {"id": "java-default-v1", "version": 3},
  "source": {
    "kind": "object",
    "artifact_id": "uuid",
    "revision": "commit-or-content-sha"
  },
  "target": null,
  "limits": {
    "timeout_seconds": 7200,
    "max_cpu": 4,
    "max_memory_mb": 8192,
    "max_output_bytes": 104857600
  },
  "data_snapshot_ids": ["osv-2026-10-01"]
}
```

平台不传递不必要的用户身份、Cookie 或完整凭证。所有凭据通过短时引用/挂载，并在任务结束后失效。

当 Runner 与引擎位于同一主机或共享卷可达的部署单元时，`source` SHOULD 使用任务级工作目录引用，不应通过控制 HTTP/RPC 上传源码或大型镜像。Runner 必须根据固定根目录和 `scan_id` 推导路径，拒绝任意绝对路径、`..` 和符号链接逃逸。

## 3. 进程与状态

Runner 向平台或本地协调器发送事件：

- `accepted`：输入校验完成；
- `started`：引擎已启动；
- `progress`：单调 stage + 可选百分比，不伪造完成度；
- `heartbeat`：证明租约存活；
- `diagnostic`：非致命输入/引擎问题；
- `completed`：包含状态和产物清单；
- `failed`：稳定错误码和脱敏消息；
- `canceled`：已终止所有子进程并完成清理。

错误码至少覆盖：`INVALID_INPUT`、`UNAUTHORIZED`、`TARGET_UNREACHABLE`、`ENGINE_UNAVAILABLE`、`ENGINE_FAILED`、`TIMEOUT`、`RESOURCE_EXHAUSTED`、`OUTPUT_INVALID`、`CANCELED`。

## 4. 结果批次

- 大结果采用 NDJSON 或分页批次；
- 每批包含 `scan_id`、`attempt`、`batch_sequence` 和 `content_sha256`；
- 平台以 `(scan_id, attempt, batch_sequence)` 幂等接收；
- 完成消息前需要提交所有批次和最终清单；
- 批次乱序时按序号或 manifest 组装，不能以到达顺序推断完整性。

## 5. 统一结果

必填类型：

- `Finding`
- `Component`
- `Occurrence`
- `ArtifactReference`
- `Diagnostic`
- `EngineMetadata`

字段与枚举见 [scan-result-model.md](scan-result-model.md)。

## 6. 兼容性

- 契约使用语义版本；新增可选字段为 minor，改变语义或必填为 major；
- 平台在滚动升级期间至少兼容前后一个 minor；
- 引擎原始格式永远保留在对象存储，映射失败可离线重放；
- 适配器需声明所支持的引擎版本范围和测试过的版本。

## 7. 适配器验收

- 健康检查、正常执行、取消、超时、OOM、报告损坏；
- 重复提交批次不会生成重复 Finding；
- 不泄漏 Runner 临时路径和 Secret；
- 固定样本结果与黄金结果等价；
- 引擎版本升级有回归报告。
