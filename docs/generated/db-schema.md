# 概念数据库 Schema

状态：设计稿，非可执行迁移。生成于 2026-10-29。

## 约束

- 所有租户资源带 `tenant_id`，外键与唯一索引包含租户作用域；
- 时间统一为 UTC；
- ID 使用 UUIDv7 或等价的时序 UUID；
- 原始报告引用对象存储 key 和 SHA-256，不保存源码全文；
- PII/Secret 不进入自由文本日志字段；
- 软删除只用于业务恢复，扫描证据按保留策略物理/对象生命周期清理。

## 实体

```text
tenants(id, name, status, created_at)
users(id, tenant_id, external_subject, email, display_name, status)
projects(id, tenant_id, name, slug, source_kind, owner_team_id, status)
role_bindings(id, tenant_id, subject_type, subject_id, project_id, role, expires_at)
repositories(id, tenant_id, project_id, provider, external_id, default_branch)
scan_targets(id, tenant_id, project_id, url, environment, allowed_hosts, allowed_cidrs,
             owner_id, approver_id, valid_from, valid_to, status)
scan_profiles(id, tenant_id, project_id, type, name, version, config_json, active)
scans(id, tenant_id, project_id, type, profile_version, status, target_id,
      source_revision, input_artifact_id, requested_by, idempotency_key,
      priority, attempt, created_at, started_at, finished_at, failure_code)
scan_steps(id, tenant_id, scan_id, engine, adapter_version, status,
           started_at, finished_at, failure_code)
scan_artifacts(id, tenant_id, scan_id, kind, object_key, content_sha256,
               size_bytes, content_type, expires_at)
findings(id, tenant_id, project_id, scan_id, type, engine, rule_id,
         fingerprint, title, category, severity, confidence, status,
         first_seen_at, last_seen_at, updated_at, details_json)
occurrences(id, tenant_id, finding_id, location_json, evidence_artifact_id)
finding_events(id, tenant_id, finding_id, from_status, to_status,
               actor_id, reason, created_at)
suppressions(id, tenant_id, project_id, selector_json, scope,
             reason, approved_by, valid_until, status)
policies(id, tenant_id, project_id, name, version, config_json, active)
policy_decisions(id, tenant_id, scan_id, policy_version, decision, reasons_json)
webhook_subscriptions(id, tenant_id, project_id, endpoint, secret_ref, events_json, status)
webhook_deliveries(id, tenant_id, subscription_id, event_id, status,
                   attempt, next_retry_at, response_status)
outbox_events(id, tenant_id, event_type, aggregate_id, payload_json,
              created_at, published_at)
audit_events(id, tenant_id, actor_type, actor_id, action, resource_type,
             resource_id, result, request_id, ip_hash, metadata_json, created_at)
data_snapshots(id, source, version, fetched_at, content_sha256, status)
data_sync_runs(id, source, started_at, finished_at, cursor, status, error_code)
iast_environments(id, tenant_id, project_id, app_version, runtime, status,
                  valid_from, valid_until, approver_id)
iast_agents(id, tenant_id, environment_id, agent_version, status,
            last_heartbeat_at, revoked_at)
quota_usage(id, tenant_id, project_id, period_start, scans_used,
            cpu_seconds, storage_bytes)
```

## 关键索引

- `scans(status, priority, created_at)`：调度队列；
- `findings(tenant_id, project_id, status, severity, last_seen_at)`：风险列表；
- `findings(tenant_id, project_id, fingerprint)`：唯一去重；
- `occurrences(finding_id)`；
- `audit_events(tenant_id, created_at)` 和时间分区；
- `outbox_events(published_at, created_at)`：待投递事件；
- `iast_agents(status, last_heartbeat_at)`：失联回收。

## 任务租约

执行中的 `scan_steps` 或独立 `job_leases` 保存 worker ID、lease_expires_at、heartbeat_at。领取时使用事务 + `FOR UPDATE SKIP LOCKED`，状态变化与审计/outbox 在同一事务提交。

## 迁移纪律

- 所有迁移向前执行并有回滚/补偿方案；
- 大表索引异步/分批创建；
- 字段删除采用 expand/migrate/contract；
- schema 变更更新本文件和 API/SARIF 映射；
- 未经真实查询模式验证不提前增加大量索引。
