# 身份与权限

状态：草案。具体身份提供方待 Q-002 确认。

## 身份

- 用户：OIDC Authorization Code + PKCE，后端 BFF 管理中短时会话；
- 服务账户：项目级机器令牌，限定 scope 和有效期；
- Runner：独立服务身份，只可领取/更新自己任务及写入独占前缀；
- IAST Agent：一次性注册换取短期身份，只能上传关联会话结果。

不在浏览器保存扫描器 API key；不在 Runner 环境变量中放置平台数据库凭据。

## 权限动作

```text
project:read
scan:create
scan:cancel
finding:read
finding:triage
finding:accept_risk
target:manage
target:approve
iast:deploy
policy:manage
member:manage
audit:read
report:download_raw
```

## 角色到动作

| 动作 | viewer | developer | security_reviewer | admin |
|---|---:|---:|---:|---:|
| 查看项目/结果 | 是 | 是 | 是 | 是 |
| 创建/取消扫描 | 否 | 是 | 是 | 是 |
| 处理/指派问题 | 否 | 是 | 是 | 是 |
| 接受风险 | 否 | 否 | 是 | 是 |
| 目标登记 | 否 | 否 | 是 | 是 |
| 目标审批 | 否 | 否 | 是 | 是 |
| IAST 部署 | 否 | 否 | 是 | 是 |
| 策略/成员管理 | 否 | 否 | 否 | 是 |
| 下载原始证据 | 否 | 否 | 是 | 是 |

职责分离要求：目标登记人与审批人建议不同；接受高风险需要第二审批（待 Q-013 确认）。

## 授权规则

- 每个查询带租户和项目作用域；
- 用户可在多个项目有不同角色，取并集但不跨项目；
- 服务账户按项目、scope 和 CIDR（可选）限制；
- 禁止通过猜测 UUID 访问资源；
- 下载短时 URL 生成时重新校验权限；
- 权限变更写审计并立即失效授权缓存；
- IAST/DAST 高风险操作要求资源当前有效、未过期且未禁用。
