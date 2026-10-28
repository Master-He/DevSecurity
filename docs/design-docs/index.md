# 设计文档索引

设计文档记录“为什么这样设计”以及稳定接口，不重复产品需求。

| 文档 | 状态 | 内容 |
|---|---|---|
| [core-beliefs.md](core-beliefs.md) | 已采用 | 工程与产品判断原则 |
| [engine-adapter-contract.md](engine-adapter-contract.md) | 草案 | Runner 与平台的控制/结果契约 |
| [scan-result-model.md](scan-result-model.md) | 草案 | 统一 Finding、SBOM、证据模型 |
| [identity-and-permissions.md](identity-and-permissions.md) | 草案 | 身份、角色和资源授权 |

## 待新增设计

- 决策记录模板与实例；
- 数据源同步快照格式；
- 对象存储生命周期与删除证明；
- DAST 出口代理实现选择；
- IAST Agent 协议和安全认证；
- 多引擎聚类算法；
- 前端设计 token 和页面原型。

新增设计必须写清楚状态、背景、决策、替代方案、风险和影响面。涉及 `QUESTION.md` 的条目要链接 ID。
