# 统一扫描结果模型

状态：草案。以下为逻辑字段，最终 JSON Schema 在实现时生成并纳入契约测试。

## Finding

| 字段 | 必填 | 说明 |
|---|---|---|
| `id` | 是 | 平台生成 UUID，不作为跨环境去重 ID |
| `fingerprint` | 是 | 项目内稳定去重指纹 |
| `tenant_id` / `project_id` | 是 | 隔离边界 |
| `scan_id` / `engine` | 是 | 来源任务和适配器 |
| `type` | 是 | `sca/sast/dast/iast` |
| `rule_id` | 是 | 规则或漏洞语义标识 |
| `title` | 是 | 面向用户的简短标题 |
| `category` | 是 | 如 injection、dependency、misconfiguration |
| `severity` | 是 | `critical/high/medium/low/info` |
| `confidence` | 是 | `confirmed/high/medium/low` |
| `status` | 是 | open/fixed/false_positive/accepted/withdrawn |
| `cwe` | 否 | CWE ID 列表 |
| `location` | 否 | 代码/组件/URL 定位 |
| `remediation` | 否 | 可执行修复建议 |
| `evidence_ref` | 否 | 受控对象引用，不放大内联 |
| `occurrences` | 是 | 至少一个发生位置 |
| `created_at` / `updated_at` | 是 | UTC |

## 扩展字段

`details` 使用带 `schema` 和 `schema_version` 的对象：

```json
{
  "details": {
    "schema": "finding.sca/v1",
    "schema_version": "1.0",
    "data": {
      "purl": "pkg:maven/group/artifact@1.2.3",
      "source_record": "OSV-2026-1234",
      "dependency_path": ["root", "group:lib@1.2.3"],
      "fixed_versions": ["1.2.4"]
    }
  }
}
```

不同引擎的必填字段放在固定顶层，引擎特有数据不能污染公共字段语义。前端只读取已版本化 schema，不直接解析原始报告。

## Occurrence

- `location_kind`: `source/component/endpoint/runtime`
- `file_path`、`start_line`、`end_line`、`column`
- SCA：`purl`、依赖路径；
- DAST：规范化 endpoint、method、参数类别；
- IAST：服务、路由、source/sink、call chain；
- 所有位置禁止包含 Runner 绝对路径。

## 去重

建议指纹输入：

- SCA：项目 + ecosystem + package + source record + affected range + direct/transitive；
- SAST：项目 + 规则语义 + 根因位置 + 数据流端点，不以消息文本为主；
- DAST：项目 + 规则 + 规范化 endpoint + 参数 + 根因；
- IAST：项目 + 规则 + endpoint + 参数 + sink + 调用栈根因。

同一指纹可有多个 occurrence。改变规则版本或根因应允许生成新 Finding，并在 UI 展示关联，不能粗暴覆盖历史。

## SBOM

`Component`：

- `id`、`purl`、`ecosystem`、名称、版本；
- `scope`: runtime/dev/build/optional；
- `direct`、`dependency_path`；
- `licenses`、`source_refs`；
- `provenance` 和原始清单位置。

SBOM 默认 CycloneDX JSON；Finding 与 SBOM 的组件 ID 需稳定映射。SPDX 作为后续导出格式，不作为首版内部模型。

## 状态与安全

- `status` 是平台级处置状态，不等于引擎原始状态；
- `confirmed` 置信度也不自动等于人工已验证；
- 接受风险必须关联审批、理由和到期时间；
- 证据引用需要对象权限，列表 API 不返回源码全文；
- 所有外部字符串按不可信输入处理。
