# Platform Backend SPEC

## 1. 职责

提供身份、项目、扫描编排、结果管理、策略、审计和对外 API。后端是唯一写入任务状态与结论的服务，Runner 只能通过专用接口提交任务结果。

## 2. 技术基线

- Go 1.24+，标准库 HTTP 路由或轻量路由器，具体框架在实现 benchmark 后确定；
- PostgreSQL 16+，迁移工具在实现时锁定；
- OpenAPI 3.1 作为 API 契约；
- S3 兼容对象存储；
- OIDC/OAuth 2.1 BFF 认证，LDAP 作为可选身份导入；
- Prometheus 指标、结构化日志和审计事件；
- API 版本 `/api/v1`，异步长任务。

不使用 PostgreSQL 保存源码全文；数据库保存元数据、结果和对象引用。任务调度使用数据库租约，避免首版引入额外消息系统。

## 3. 领域模型

核心表见 [../../docs/generated/db-schema.md](../../docs/generated/db-schema.md)。

### Scan

- `id`, `tenant_id`, `project_id`, `type`, `profile_id`;
- `source_revision`, `input_artifact_id`, `target_id`;
- `requested_by`, `status`, `priority`;
- `created_at`, `started_at`, `finished_at`, `expires_at`;
- `idempotency_key`, `attempt`, `failure_code`;
- 每个任务最多一个 active run，重试使用新 attempt 并保留历史。

### Finding

- 统一字段：ID、项目、扫描、引擎、规则、类别、严重性、置信度、状态、位置、CWE、修复建议；
- 引擎专属字段放 JSONB，不把不同引擎的必填字段硬编进公共模型；
- `fingerprint` 项目内稳定，去重可解释；
- 原始证据作为受控对象，不嵌入列表 API。

### 权限

角色建议：

- `viewer`：查看项目结果；
- `developer`：创建扫描、处理分配给自己的问题；
- `security_reviewer`：复核、标记误报、接受风险和配置策略；
- `admin`：平台配置、用户和集成；
- `service`：限定项目、动作和有效期的机器身份。

权限采用资源级 RBAC，不能只依赖前端隐藏按钮。

## 4. API 契约

### 上传源码

```http
POST /api/v1/projects/{project_id}/source-uploads
```

请求声明文件名、大小和 SHA-256 的预期值；响应返回短时上传地址、`upload_id` 和过期时间。上传完成后调用：

```http
POST /api/v1/projects/{project_id}/source-uploads/{upload_id}/complete
```

平台验证大小、哈希、压缩包结构和路径安全后，才允许该产物创建 SCA/SAST 扫描。上传未完成或校验失败的对象在短时窗口后删除。

### 创建扫描

```http
POST /api/v1/projects/{project_id}/scans
Authorization: Bearer <service-token-or-session>
Idempotency-Key: <unique-key>
Content-Type: application/json
```

```json
{
  "type": "sca",
  "profile_id": "java-default-v1",
  "source": {
    "kind": "git",
    "revision": "40-hex-sha",
    "repository_id": "repo-id"
  },
  "metadata": {
    "pipeline_run_id": "ci-123",
    "branch": "main"
  }
}
```

响应 `202 Accepted`，返回 `scan_id` 和 `status_url`。输入必须是支持的 kind，禁止把任意 URL 直接当作 Git endpoint。

### 查询与结果

```http
GET /api/v1/scans/{scan_id}
GET /api/v1/scans/{scan_id}/findings?cursor=<opaque>&limit=100
GET /api/v1/findings/{finding_id}
GET /api/v1/scans/{scan_id}/report.sarif
POST /api/v1/scans/{scan_id}/cancel
```

结果分页使用 cursor，排序稳定。列表只返回摘要；证据和原始报告需要额外权限及短时签名 URL。

### 机器集成

- 服务令牌按项目、动作和过期时间授权；令牌仅展示一次、存储哈希。
- Webhook 使用 HMAC-SHA256 签名、时间戳和重放窗口；接收方返回 2xx 才确认投递。
- 失败重试使用指数退避，超过次数进入 dead-letter 状态，可在 UI 查看和重放。
- 所有写操作带 `request_id`；计划幂等，查询和取消天然幂等。

错误格式：

```json
{
  "error": {
    "code": "TARGET_NOT_AUTHORIZED",
    "message": "目标不在允许范围内",
    "request_id": "uuid",
    "details": {}
  }
}
```

不向未授权调用者暴露目标解析地址、内部主机名、源码路径或扫描器细节。

## 5. 扫描编排

### 状态机

```text
queued -> provisioning -> running -> finalizing
   |           |             |           |
   +-----------+-------------+-----------+-> failed
   +-------------------------------------> canceled
   +-------------------------------------> expired
running/finalizing -> partially_succeeded
```

- 非法状态转变拒绝并记录原因。
- Worker 领取任务时设置租约和心跳；失联按策略重新排队。
- Runner 结果提交以 `attempt + batch_sequence` 幂等。
- 引擎不可用不默认变成 0 个漏洞；必须标为诊断/部分成功。

### 源码与对象存储

- 上传使用短时、限大小、内容类型约束的方式。
- 对象以租户和扫描隔离；Runner 只能写自己的前缀。
- 源码和原始报告按 Q-009 自动过期；删除任务写审计并异步清理对象。
- 前端下载使用短时签名 URL，并再次校验项目权限。

## 6. 去重与策略

- 去重发生在后端，不能只依赖扫描器输出。
- 去重键包含项目、规则语义、路径/参数、组件/依赖和根因；保留所有 occurrence。
- 策略计算输入：风险、置信度、传播关系、项目等级和例外。
- 门禁结论必须包含策略版本、输入结果和决策原因。
- 接受风险必须有到期日、审批人和复核流程；到期自动重新打开。

## 7. 审计与隐私

记录以下事件：登录/登出、授权变更、项目/目标变更、Secret 使用、扫描创建/取消/重试、结果下载、状态变更、策略变更、Webhook 重放。

审计事件包含操作者、动作、资源、结果、时间、request ID；不能记录完整 Secret、Cookie 或源码。

## 8. 非功能需求

- API 99.9% 月可用性目标（待 Q-011 确认）。
- 常规查询 P95 < 500 ms，创建扫描 < 1 s（不含扫描执行）。
- 支持至少 300 用户、80 并发 UI、每天 1,000 次扫描的规划基线。
- 任务、结果和审计都支持游标翻页。
- 数据库迁移支持向前和回滚计划；发布不丢失进行中的任务。
- 所有关键错误可关联 request ID 和运维指标。

## 9. 测试

- OpenAPI schema 与实现契约测试；
- 数据库租约、重复领取、重试、取消和 Runner 崩溃模拟；
- 幂等创建和重复结果提交；
- RBAC、跨项目访问、对象签名 URL 越权测试；
- Webhook 签名、重放和死信测试；
- Finding 去重稳定性和策略解释测试；
- 大分页、超大报告、恶意文件名和归档测试。

## 10. 实施顺序

1. 认证、项目、审计、对象存储；
2. Scan/Job 状态机与一个假引擎适配器；
3. Finding 模型、游标查询、SARIF 输出；
4. SCA/SAST Runner 接入；
5. DAST/IAST 受控适配；
6. 策略、Webhook、报表和容量调优。

本文件是后端 API 与领域行为规格；数据库最终 schema 以生成文件对应的迁移为准。
