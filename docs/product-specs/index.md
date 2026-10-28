# 产品规格索引

## 产品定位

为内网几百人规模公司提供统一的安全 SDLC 控制面，通过 API 集成已有 DevOps，以 SCA/SAST/DAST/IAST 覆盖代码、组件、运行测试和交互测试四个阶段。

## 范围

首版必须完成：

- 项目、成员、角色和审计；
- 扫描任务创建、排队、执行、取消、失败和部分成功；
- 统一风险列表、详情、处置和导出；
- CI/API 集成与 Webhook；
- SCA/SAST 主语言闭环；
- DAST/IAST 的受控试点能力。

首版不需要：

- 自研所有语言 parser；
- 无限制漏洞网站爬取；
- 未授权目标扫描；
- 生产环境 IAST；
- 自动修复或自动合并代码。

## 产品规格

| 文档 | 受众 | 说明 |
|---|---|---|
| [new-user-onboarding.md](new-user-onboarding.md) | 研发/安全/DevOps | 首次项目和 CI 接入 |
| [scan-workflow.md](scan-workflow.md) | 平台/扫描器开发者 | 任务与失败状态 |
| [finding-lifecycle.md](finding-lifecycle.md) | 研发/安全 | 风险处置和验收 |
| [platform-integration.md](platform-integration.md) | DevOps | API、幂等、Webhook 和门禁 |

## 通用验收原则

- 未发现问题和扫描未完成必须可区分；
- 所有结果都有来源、版本和定位；
- 所有高风险操作都有权限和审计；
- 项目之间的数据默认隔离；
- 错误信息不泄漏内部地址、Secret 或源码。
