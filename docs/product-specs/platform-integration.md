# DevOps 平台集成

## 集成模式

1. **同步创建、异步查询**：CI 创建扫描后拿到 ID，轮询/Webhook 等待结果。
2. **上传源码**：无 Git 连接权限时，CI 上传单次源码归档；不使用任意 URL。
3. **Git 适配器**：平台使用受控集成拉取固定 revision，不让用户提供随机凭据。
4. **SARIF**：向代码托管平台上传标准结果，但平台自身 API 保留完整模型。

## 安全令牌

- 服务令牌绑定组织/项目、scope、有效期和创建人；
- 明文只显示一次，数据库仅保存哈希；
- CI Secret 不进 URL、日志和构建产物；
- 建议每个流水线使用独立令牌，便于吊销和审计。

源码上传采用两阶段流程：先创建短时 `source-upload`，再上传归档并调用 `complete`。只有校验大小、SHA-256、归档结构和路径安全后，扫描请求才能引用该产物。

## 创建与等待

```bash
curl -X POST "$DEVSECURITY_URL/api/v1/projects/$PROJECT_ID/scans" \
  -H "Authorization: Bearer $DEVSECURITY_TOKEN" \
  -H "Idempotency-Key: ${CI_PIPELINE_ID}-${CI_COMMIT_SHA}-sca" \
  -H "Content-Type: application/json" \
  -d '{"type":"sca","profile_id":"java-default-v1","source":{"kind":"git","repository_id":"repo-id","revision":"'$CI_COMMIT_SHA'"},"metadata":{"pipeline_run_id":"'$CI_PIPELINE_ID'"}}'
```

CI 应保存 `scan_id`，在退出前按项目可接受时限轮询；超时不等同于扫描失败，需保留结果链接或回调处理。

## 轮询与 Webhook

- 轮询建议从 10 秒开始，指数退避到 60 秒；
- Webhook 带事件 ID、时间戳、签名和版本；
- 接收方按事件 ID 幂等，校验重放窗口；
- 门禁读取策略决策，而不是自己把所有 Critical 都作为同一规则；
- 接收方失败不影响扫描结果保存，平台支持重试与死信。

## 门禁决策

建议响应：

```json
{
  "scan_id": "uuid",
  "status": "succeeded",
  "decision": "warn",
  "policy": {"id": "default-v1", "version": 4},
  "reasons": ["发现 2 个 high 问题，项目默认为警告"],
  "report_url": "/api/v1/scans/uuid/report.sarif"
}
```

`decision` 建议为 `pass/warn/block/error`。 `error` 必须与无问题区分；默认不因扫描基础设施故障自动阻断生产，是否阻断由 Q-012 决定。

## 兼容性与审计

- API 版本与响应 schema 版本分离；
- breaking change 至少提前一个稳定版本告知；弃用有日期；
- 每个请求带 `request_id`，CI 日志可关联；
- 查询和导出按项目权限审计；
- 同一个幂等键不能跨项目复用。
