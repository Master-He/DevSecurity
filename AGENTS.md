# AGENTS.md

本文件是仓库内所有开发者和自动化 Agent 的入口。修改代码或文档前先阅读相关子目录说明；若子目录后续新增 `AGENTS.md`，其规则优先于本文件。

## 当前状态

仓库已进入可运行 PoC 阶段：

- `platform/backend`：Spring Cloud 网关、Eureka 和控制面服务；
- `platform/frontend`：React 管理界面；
- `SCA`：可运行的 Go Maven SCA；
- `SAST/joern-adapter`：Joern CPG 接入骨架；
- `DAST`、`IAST`：仍然只允许维护规格，不实现扫描代码。

## 不可违反的边界

1. DAST 只能扫描经过授权、登记在目标白名单内的系统。禁止把平台做成任意 URL 的公开 SSRF 代理。
2. IAST Agent 只允许安装到已授权的短期测试环境，默认禁止生产环境注入；必须提供一键卸载、超时回收和审计。
3. SCA 数据采集必须遵守数据源许可证、服务条款和访问频率限制。不得将未知许可证的数据源直接纳入商业分发。
4. 扫描器不得执行不可信仓库中的构建脚本来“识别依赖”。需要构建时必须在隔离、无凭据、默认断网的 Runner 中运行。
5. 源代码、制品、访问令牌、Cookie 和扫描证据均属于高敏数据。日志中不得出现完整令牌、Cookie、私钥或源代码大段内容。
6. 外部 Git、漏洞库、通知和身份提供方都必须通过受控适配器接入；核心业务逻辑不得依赖单个供应商。

## 文档优先级

- 系统边界和组件关系：[ARCHITECTURE.md](ARCHITECTURE.md)
- 产品行为和验收条件：[docs/product-specs/index.md](docs/product-specs/index.md)
- 引擎能力与接口：[Scanners](#扫描引擎) 和 [docs/design-docs/engine-adapter-contract.md](docs/design-docs/engine-adapter-contract.md)
- 未决业务问题：[QUESTION.md](QUESTION.md)
- 推进顺序：[docs/PLANS.md](docs/PLANS.md)

### 扫描引擎

- [SCA/SPEC.md](SCA/SPEC.md)
- [SAST/SPEC.md](SAST/SPEC.md)
- [DAST/SPEC.md](DAST/SPEC.md)
- [IAST/SPEC.md](IAST/SPEC.md)

## 设计原则

- 先定义可验证的接口和结果模型，再选具体引擎。
- 优先组合成熟开源引擎并通过适配器隔离，而不是一次性自研多语言规则。
- 低置信度结果必须显式标注，不把扫描器的“发现”直接等同于已确认漏洞。
- 业务状态以 PostgreSQL 中的任务和结果为准，扫描日志和原始报告只作为可追溯证据。
- 所有对外操作需要身份、项目级授权、审计记录和幂等语义。
- 文档中的容量、SLO、语言范围和许可证是待验证假设，不应伪装成已确认需求。

## 修改规则

1. 修改架构前先更新或新增决策文档，再改实现。
2. 新增扫描器适配器必须实现统一契约、结果去重和至少一个固定样本回归测试。
3. 修改 API 必须同步 OpenAPI、集成说明和兼容性策略。
4. 不得在仓库中提交真实源代码样本、真实漏洞数据、生产凭据或个人数据。
5. 文档中的“必须/MUST”表示验收条件，“应该/SHOULD”表示推荐实现；“计划/建议”不代表已实现。

## 完成门禁

实现阶段每个功能至少需要：

- 对应产品的验收场景；
- 权限、审计和失败路径；
- 固定样本或契约测试；
- 可观测指标和运行手册；
- 文档与 `QUESTION.md` 状态同步。

## 构建与验证

- 后端：`cd platform/backend && mvn test`
- 前端：`cd platform/frontend && npm run build`
- SCA：`cd SCA && go test ./... && go vet ./... && go build ./cmd/sca`
- 完整环境：从仓库根目录运行 `./scripts/dev-up.ps1` 或 `./scripts/dev-up.sh`
- 本地无 Docker 的 API 验证使用 Spring profile `local`，不得把 H2/文件存储当成生产配置

提交前必须运行受影响模块的构建；涉及扫描结果映射时还要运行固定样例回归。Docker Compose 完整启动需在 Docker Engine 可用时验证。
