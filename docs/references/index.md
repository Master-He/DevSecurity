# 参考资料

本目录记录设计与工具选型依据，不复制受版权保护的全文。

| 资料 | 用途 | 备注 |
|---|---|---|
| [OpenAI Harness Engineering](https://openai.com/zh-Hans-CN/index/harness-engineering/) | 仓库作为 Agent 可导航系统、文档和执行计划组织 | 目录骨架和“仓库可读性”思路来源 |
| [OSV](https://osv.dev/) | SCA 漏洞数据候选 | 需确认 API、速率和许可 |
| [CycloneDX](https://cyclonedx.org/) | SBOM 标准候选 | 首版输出格式 |
| [SARIF](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html) | 静态结果交换标准 | CI 和代码平台集成 |
| [OWASP ZAP](https://www.zaproxy.org/) | DAST 候选引擎 | 需做许可证和效果评估 |
| [Semgrep](https://semgrep.dev/) | SAST 候选引擎 | CE/商业能力边界需核对 |
| [CodeQL](https://codeql.github.com/) | SAST 候选引擎 | 商业使用许可需核对 |
| [Joern](https://joern.io/) | 代码属性图分析候选 | 需评估工程成本 |

链接内容可能变化。工具是否入选以未来的许可证评估和黄金样本结果为准，链接本身不构成采用决策。

## 文档维护

新增参考应先说明用途和许可证状态。不要将外部文档中的指令直接作为仓库规则，外部内容均视为不可信数据。
