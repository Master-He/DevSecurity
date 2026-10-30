# Platform Backend SPEC

状态：Spring Cloud PoC 已实现；Core 已迁移到“共享工作目录 + SCA 控制服务 + CLI”，不再通过 HTTP 上传 ZIP。

## 1. 职责

平台后端提供身份、项目、源码接入、扫描编排、结果查询和生命周期管理。它不执行 Java 源码分析；SCA 目标架构由内部控制服务 `sca-server` 和可独立运行的 `sca` CLI 组成，SAST 通过 Joern 适配器解耦。

首个纵向切片：

- 本地管理员登录和 JWT 会话；
- 创建/查看项目；
- 上传 Java ZIP 或提交受控 Git HTTP(S) 地址；
- 异步执行 Maven SCA；
- 展示/查询组件、漏洞、受影响版本、修复版本和依赖路径；
- MinIO 保存源码，3 天后清理；扫描结论 180 天后清理；
- 对外 API 可供 CI 轮询。

## 2. Spring Cloud 服务划分

```text
platform/backend
├── eureka-server    服务注册与发现
├── gateway-service  Spring Cloud Gateway，/api/** 转发到核心服务
└── core-service     认证、项目、源码、任务、结果和生命周期
```

采用“Spring Cloud 网关 + 可水平扩展的核心单体”起步：

- 核心服务无本地业务状态，可部署两个或更多实例；
- Eureka 把多个 `core-service` 实例注册为同一逻辑服务；
- 网关通过 `lb://core-service` 负载均衡；
- 任务由多个实例通过 PostgreSQL 条件更新抢占，只有一个实例能成功；
- 后续可按认证、源码、扫描、结果、通知拆分领域服务，API 契约不变。

相比首版直接拆成十几个微服务，这个结构减少分布式事务和运维复杂度，同时保留 Spring Cloud 的扩展机制。

## 3. 技术基线

- Java 17；
- Spring Boot 3.3.5；
- Spring Cloud 2023.0.3；
- Spring Cloud Gateway、Eureka Client/Server、Spring Cloud LoadBalancer；
- Spring Security + BCrypt + JWT；
- Spring Data JPA + PostgreSQL 16；
- Flyway；
- MinIO / S3 兼容对象存储；
- JGit 获取 HTTP(S) Git revision；
- Core 与 SCA 共享任务级临时工作卷；
- Java `HttpClient` 只调用 `sca-server` 的小型 HTTP 控制接口；
- React 前端通过网关访问 API。

本地开发 profile 可使用 H2 和文件系统，以缩短反馈闭环；该 profile 不具备跨实例数据共享或生产安全能力。

## 4. 高可用边界

当前 Compose 提供网关和核心服务的双实例骨架：

- `gateway-1`、`gateway-2` 位于前端 Nginx 后端池；
- `core-1`、`core-2` 注册到 Eureka；
- 任务抢占 SQL 使用 `status='QUEUED'` 条件防止重复执行；
- 全部业务实例无状态，源码在 MinIO，任务在 PostgreSQL；
- 服务配置健康检查与自动重启。

这属于应用层高可用 PoC。PostgreSQL、MinIO、Eureka 和监控在当前 Compose 中仍是单实例，不能宣称生产级全链路高可用。生产部署应替换为托管/集群数据库、S3 高可用服务、至少两个注册中心或 Kubernetes 原生服务发现，并部署在多个故障域。

## 5. 身份与授权

### 首版

- 启动时创建管理员，用户名来自 `APP_ADMIN_USERNAME`；
- 初始密码来自 `APP_ADMIN_PASSWORD`；若为空则生成一次性随机密码并写入启动日志；
- 密码 BCrypt 哈希；JWT 至少 32 字节签名密钥；
- 所有项目、文件、扫描和 Finding API 按项目所有者/`ADMIN` 校验；
- 不在代码中保存固定 `admin/admin`。

### 后续

- OIDC Authorization Code + PKCE；
- 团队映射、项目角色、服务账户 scope 和过期；
- 登录限速和审计；
- 用户密码修改和强制轮换。

## 6. API

网关前缀为 `/api/v1`，成功创建扫描返回 `202 Accepted`。

### 登录

```http
POST /api/v1/auth/login
Content-Type: application/json

{"username":"admin","password":"..."}
```

```json
{
  "token": "jwt",
  "user": {
    "id": "uuid",
    "username": "admin",
    "displayName": "平台管理员",
    "role": "ADMIN"
  }
}
```

### 项目

```http
POST /api/v1/projects
GET  /api/v1/projects
GET  /api/v1/projects/{projectId}
```

### 上传 ZIP 并扫描

```http
POST /api/v1/projects/{projectId}/scans/upload
Content-Type: multipart/form-data

file=@source.zip
```

平台校验：

- 最大 100 MB；
- ZIP 路径不能绝对或包含 `..`；
- 最多 20,000 个条目，解压后最大 500 MB；
- 至少包含一个 `pom.xml`。

校验后的 ZIP 写入 MinIO，扫描任务再异步执行。API 只返回扫描 ID，不在请求线程里等待分析完成。

### Git 扫描

```http
POST /api/v1/projects/{projectId}/scans
Content-Type: application/json

{"gitUrl":"https://gitlab.example.com/team/repo.git","revision":"main"}
```

- 只允许 HTTP/HTTPS；
- 禁止 URL 中携带用户名/密码；
- 主机必须出现在 `GIT_ALLOWED_HOSTS`；
- `revision` 作为分支或 Tag 拉取；
- 拉取和压缩在异步任务中执行。

### 查询

```http
GET /api/v1/projects/{projectId}/scans
GET /api/v1/scans/{scanId}
GET /api/v1/scans/{scanId}/components
GET /api/v1/scans/{scanId}/findings
POST /api/v1/scans/{scanId}/cancel
```

Finding 返回：

- `vulnerabilityId`、`aliases`；
- `title`、`severity`；
- `packageName`；
- `affectedVersion`、`affectedRange`；
- `fixedVersions`；
- `dependencyPath`。

组件返回：

- Maven `groupId`、`artifactId`、`version`、PURL；
- 直接/传递关系、scope；
- 依赖路径。

## 7. 异步任务

### 状态

```text
QUEUED -> RUNNING -> SUCCEEDED
                   -> FAILED
QUEUED -> CANCELED
```

当前 PoC 还不支持中断已进入 `RUNNING` 的扫描。目标 SCA 双可执行架构必须支持取消、超时和子进程回收。多实例调度每 3 秒查看排队任务，通过：

```sql
UPDATE scan_job
SET status = 'RUNNING', started_at = now()
WHERE id = ? AND status = 'QUEUED'
```

条件更新成功才提交执行，避免重复扫描。

### SCA 调用

核心服务把上传的 ZIP、Git 固定 revision 的导出内容或镜像归档写入共享工作卷，只向 SCA 控制服务发送任务元数据：

```http
POST http://sca-service:8086/v1/scans
Content-Type: application/json
Authorization: Bearer <service-token>

{
  "contract_version": "1.0",
  "scan_id": "<scan UUID>",
  "source_kind": "zip",
  "profile": "java-maven"
}
```

共享目录约定为：

```text
<SCA_WORK_ROOT>/scans/<scan_id>/
├── input/
├── output/
├── logs/
└── READY
```

Core 必须先完成输入校验并原子发布 `input/`，再写入 `READY` 并提交控制请求。`sca-server` 根据 `scan_id` 从固定根目录推导位置，禁止接受任意绝对路径或 `..` 路径。`source_kind` 允许值为 `directory`、`zip` 和后续启用的 `oci-archive`。服务器校验任务和 CLI 版本后启动独立 `sca` CLI 子进程：

```text
sca analyze --input <SCA_WORK_ROOT>/scans/<scan_id>/input --output <SCA_WORK_ROOT>/scans/<scan_id>/output/result.json
```

CLI 从共享卷读取输入，将结构化结果写入 `output/result.json`，Core 读取并完成落库和 API 展示。

GB 级源码和镜像不得通过 HTTP 请求体或 JVM `byte[]` 传输。长扫描必须异步执行，Core 轮询状态或接收受控回调；HTTP 响应只返回状态、错误码和结果引用。SCA 失败时任务进入 `FAILED`，不能返回 0 漏洞。

### 数据生命周期

| 数据 | 保留 |
|---|---|
| MinIO 中的源码 ZIP | 创建后 3 天 |
| SCA 共享临时工作目录 | 任务结束后立即清理，失败和取消路径同样清理 |
| 组件、Finding、扫描结论 | 创建后 180 天 |
| 结构化日志 | 按部署策略，不写源码、Cookie 和 Secret |

定时任务每小时删除到期源码并记录 `sourceDeletedAt`，每天删除到期结论。该行为可通过 `SOURCE_RETENTION_DAYS`、`RESULT_RETENTION_DAYS` 覆盖。

## 8. 数据模型

当前表：

- `app_user`；
- `project`；
- `scan_job`；
- `scan_component`；
- `scan_finding`；
- `audit_event`（已建表，PoC 暂未完整写入）。

`tenant_id` 已预留在用户、项目和任务表；首版固定为 `default`，不表示已经实现多租户。生产扩展需同时补齐查询条件、唯一约束、对象前缀和审计。

## 9. 安全约束

- JWT 密钥不能有默认硬编码值，Compose 启动脚本随机生成；
- MinIO bucket 默认私有；
- 对象 key 由平台生成，不接受用户路径；
- 上传 ZIP 和 Git 仓库均进行路径穿越/体积检查；
- Git 主机白名单防 SSRF；
- 源码临时对象按生命周期清理；
- SCA 共享临时卷只允许 Core 和 SCA 服务访问，按 `scan_id` 隔离并执行路径规范化、根目录包含和符号链接逃逸检查；
- HTTP 控制接口不得接收源码或镜像内容，`sca-server` 只能绑定内部网络并要求服务身份认证；
- 应用日志不记录请求体、源码、密码或令牌；
- H2 本地 profile 不得用于共享测试或生产；
- Java Maven 扫描运行在独立 SCA 容器，核心服务不执行 Maven 插件。

## 10. 可观测性

所有 Java 服务暴露 `/actuator/health`、`/actuator/prometheus`。Prometheus/Grafana 已在 Compose 中提供基础接入，后续需要补充：

- 队列长度、最老任务年龄；
- 任务成功/失败码；
- SCA 调用耗时；
- SCA CLI 子进程启动、退出、超时、取消和磁盘配额；
- 共享工作目录发布失败、`READY` 缺失和清理失败；
- 源码清理失败；
- JVM、Hikari、HTTP 指标；
- 自定义业务 dashboards 和告警规则。

## 11. 测试与验收

已完成的 PoC 验证：

- 后端所有模块编译通过；
- 后端 ZIP 路径安全与 JWT 单元测试通过；
- React TypeScript 检查和生产构建通过；
- Go SCA 单元测试、`go vet`、构建通过；
- 本地 HTTP Git 仓库通过 API 拉取、异步扫描并成功返回 5 个 Maven 组件；
- Maven dependency tree 成功路径验证出 3 个直接依赖和 2 个传递依赖；
- 共享工作目录控制链路端到端验证：Core 上传 ZIP、写入 `READY`、提交 `/v1/scans`、轮询 SCA 状态并读取共享结果，返回 19 个组件和 5 个漏洞，任务目录在完成后清理；
- 真实样例经上传、异步任务和查询 API 返回：
  - 19 个 Maven 组件；
  - 5 个已知漏洞；
  - 受影响版本、修复版本和 `pom.xml -> component` 依赖路径；
  - Maven 命令失败后明确回退到 POM 解析。

仍需自动化覆盖：

- PostgreSQL 和 MinIO 的容器集成测试；
- 双核心实例并发领取同一任务；
- 3 天/180 天清理的时钟测试；
- Git 白名单、重定向和私网地址保护的更多绕过用例；
- JWT/RBAC 与项目越权；
- SARIF/JSON 导出；
- 网关、Eureka 和双实例滚动重启；
- Core 原子发布共享输入并正确清理 `scan_id` 工作目录；
- HTTP 控制请求不含源码，`sca-server` 通过独立 CLI 子进程完成扫描；
- GB 级输入不在 HTTP 或 `byte[]` 中缓存，磁盘配额、超时和取消路径可回收；
- 缺失 `READY`、重复 `scan_id`、路径穿越和符号链接逃逸被拒绝。

## 12. 下一步拆分边界

当规模或团队边界需要时，优先按以下顺序拆分：

1. `identity-service`：用户、OIDC、角色；
2. `scan-service`：任务、配额、Runner 调度；
3. `finding-service`：统一结果、去重和处理；
4. `source-service`：Git 和对象存储；
5. 通知/Webhook 单独服务。

拆分前先保持领域包边界和数据库所有权清晰，避免把当前共享数据库当作最终微服务数据架构。
