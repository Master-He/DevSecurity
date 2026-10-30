# DevSecurity

DevSecurity 是面向前端、后端和移动端研发流程的内网安全 SDLC 平台。目标是在代码提交、合并请求、制品构建和测试环境验证等环节提供统一的 SCA、SAST、DAST、IAST 扫描能力，把结果标准化后展示给研发、安全和项目负责人，并通过 API 嵌入已有 DevOps 流程。

## 当前阶段

**可运行 PoC 阶段。** 已实现 Spring Cloud 控制面骨架、React 管理界面、Go Maven SCA 和 Joern CPG 接入骨架；DAST/IAST 仍只有规格。

- `SCA/`：软件成分分析引擎规格
- `SAST/`：静态应用安全测试引擎规格
- `DAST/`：动态应用安全测试引擎规格
- `IAST/`：交互式应用安全测试引擎规格
- `platform/backend/`：Spring Cloud Gateway、Eureka 和核心服务
- `platform/frontend/`：React 管理界面
- `examples/java-vulnerable/`：用于离线演示的 Maven 样例

## 阅读顺序

1. [AGENTS.md](AGENTS.md)：仓库边界、约束和维护规则
2. [ARCHITECTURE.md](ARCHITECTURE.md)：系统架构、技术选型和关键数据流
3. [QUESTION.md](QUESTION.md)：需要业务方确认的决策
4. [docs/product-specs/index.md](docs/product-specs/index.md)：产品范围与核心流程
5. 四个扫描引擎目录中的 `SPEC.md`
6. [docs/PLANS.md](docs/PLANS.md)：阶段划分和验收标准

## 快速启动

前置条件：Docker Engine / Docker Desktop 可用。首次启动会随机生成管理员、数据库和 JWT 密码，安全配置写入未纳入 Git 的 `.env`。

Windows PowerShell：

```powershell
.\scripts\dev-up.ps1
```

Linux/macOS：

```bash
./scripts/dev-up.sh
```

启动后访问：

- Web：<http://localhost:3000>
- API 网关：<http://localhost:8080>
- MinIO 控制台：<http://localhost:9001>
- Grafana：<http://localhost:3001>
- Eureka：<http://localhost:8761>

将 `examples/java-vulnerable/` 压缩成 ZIP 后上传即可演示 Maven SCA。Git 拉取仅允许 `.env` 中 `GIT_ALLOWED_HOSTS` 登记的主机。

## 独立开发

```bash
# SCA Go 服务
cd SCA
go run ./cmd/sca analyze --input ../examples/java-vulnerable --output ./result.json

# Spring Cloud 后端
cd platform/backend
mvn test

# React 前端
cd platform/frontend
npm install
npm run dev
```

Spring `local` profile 使用 H2 和本地文件系统，仅用于无 Docker 的测试；生产/Compose 使用 PostgreSQL 和 MinIO。

## 启动排障

如果 `dev-up.ps1` 在 `Pulling` 阶段报 `failed to resolve reference`、`TLS handshake` 或 `EOF`：

1. 先执行 `docker version`，确认 `Server` 部分存在且 Engine 正在运行。
2. 执行 `docker pull postgres:16-alpine`。如果这一步也失败，说明问题在 Docker 镜像源、代理或网络，不在项目代码。
3. 检查 Docker Desktop 的 `Settings > Docker Engine > registry-mirrors`，删除已失效的加速器地址，配置当前网络可访问的源，然后点击 `Apply & restart`。
4. 检查 `Settings > Resources > Proxies`。代理端口变化或代理不可用时，Docker 拉取镜像会表现为 TLS/EOF 错误。
5. 镜像源恢复后重新运行 `.\scripts\dev-up.ps1`。

脚本现在会主动校验 Docker Engine 和 Compose 配置，并正确返回 Compose 的非零退出码。首次启动生成的本地管理员密码保存在被 Git 忽略的 `.env` 文件中的 `APP_ADMIN_PASSWORD`。

## 项目目标

- 支持几百人规模公司在内网长期使用。
- 提供统一的扫描任务、风险展示、项目视图、趋势统计和 API 集成能力。
- 允许通过适配器接入经过评估的开源扫描引擎，避免为每种语言从零实现规则体系。
- 通过统一结果模型消除不同引擎在状态、严重性、定位和去重方式上的差异。
- 对代码、扫描产物、凭据和被扫描目标实施最小权限与审计。

## 明确不在本次范围

- 自动攻击互联网上的任意目标。
- 在未授权系统中安装 IAST Agent。
- 在源代码中植入后门、持久化进程或绕过目标系统的安全控制。
- 在未确认漏洞库、工具许可证和数据保留规则前，进行大规模爬取或数据处理。

## 成功标准

首个可交付版本应能让一条流水线完成“提交代码或连接仓库 -> 选择扫描类型 -> 异步执行 -> 查看标准化结果 -> 通过 API 获取结果”的闭环，并具备可审计的权限和运行记录。具体指标见 [docs/QUALITY_SCORE.md](docs/QUALITY_SCORE.md)。

> 架构中的容量数字、SLO、默认语言优先级和部署方式是立项假设，需在 [QUESTION.md](QUESTION.md) 中确认。
