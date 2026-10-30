# SCA 引擎 SPEC

状态：Maven 分析核心、独立 CLI、HTTP 控制服务、共享工作目录校验以及 Core 侧共享卷发布/控制接口接入均已实现。

当前实现约定：

- `sca` 是不监听网络的独立 CLI，`sca-server` 是调用 CLI 的内部控制服务；
- 环境存在 `mvn` 时优先执行 `dependency:tree`；
- Maven 失败/不存在时安全回退到 XML 解析；
- 缺失或无法解析的版本统一输出 `unknown`；
- 漏洞数据以 Maven `groupId:artifactId` 和 PURL 匹配，固定样例数据仅用于演示；
- 多生态覆盖、联网数据同步和容器镜像扫描仍属于后续阶段。

本文后续将“当前已实现行为”和“完整平台目标架构”分开描述。GB 级镜像扫描和跨节点共享存储仍需后续实现。

## 1. 目的

从源代码、锁文件、构建产物和容器镜像中识别第三方组件及其依赖关系，匹配已知漏洞，输出可去重、可追溯、可采取行动的风险结果和 SBOM。

SCA 不能仅凭“存在某版本库”就断言业务可利用；结果应携带证据、置信度、直接/传递依赖关系和修复建议。

## 2. 运行形态与执行架构

### 2.1 两个可执行文件

SCA 对外提供两个逻辑可执行文件；Windows 使用 `.exe`，Linux 容器使用无后缀文件名：

| 可执行文件 | 职责 | 网络监听 |
|---|---|---|
| `sca` / `sca.exe` | 独立 CLI，读取本地源码、归档或镜像路径，执行分析并输出统一结果 | 禁止 |
| `sca-server` / `sca-server.exe` | 对可信控制面提供 HTTP 控制接口，校验任务并调用 CLI 子进程 | 仅内部控制网络 |

必须满足：

- `sca` MUST 是可独立复制和运行的命令工具，不依赖 `sca-server`、Core 或数据库。
- `sca-server` MUST NOT 复制一套分析算法；每次扫描必须调用与服务器版本匹配的 `sca` CLI。
- `sca-server` 必须使用固定可执行文件路径和参数数组启动 CLI，禁止经过 shell 拼接命令，禁止让请求内容覆盖可执行文件路径。
- SCA 容器镜像必须同时包含两个可执行文件；CLI 还必须作为独立发布产物提供，便于用户直接排障和复用。
- 两个可执行文件必须独立版本化，并通过兼容性检查阻止服务端调用不兼容的 CLI。
- HTTP 服务不得被当作扫描数据通道。源码、镜像和大型结果 MUST NOT 放入 HTTP 请求体或响应体。

CLI 的最小界面约定为：

```text
sca analyze --input <path> --output <result.json>
```

- `--input` 是本地目录、ZIP 或后续实现的容器镜像归档路径；当前 Maven-only 版本根据输入内容自动选择解析方式。
- `--output` 是统一 JSON 结果文件路径。CLI 必须原子写入该文件，避免部分结果被误认为完整报告。
- 当前版本不要求 `--profile`。以后需要选择规则集、生态或扫描策略时可增加可选参数，但不得破坏上面的最小命令。
- CLI 不得要求调用方了解 Core 的共享卷目录结构，也不得要求设置 Core 环境变量。

### 2.2 共享工作目录

Core 和 SCA 执行进程运行在同一主机、同一 Pod 或共享存储可达的部署单元中，并挂载同一个临时数据卷：

```text
<SCA_WORK_ROOT>/
└── scans/
    └── <scan_id>/
        ├── input/       # Core 写入，发布后 SCA 只读；ZIP 输入固定为 source.zip
        ├── output/      # SCA 写入结构化结果和诊断
        ├── logs/        # 脱敏后的进程日志
        └── READY        # Core 完成输入发布后的原子就绪标记
```

必须满足：

- Core 必须先把上传 ZIP、Git 导出内容或镜像归档写入 staging 目录，确认完整后原子发布为唯一的 `scan_id` 工作目录，再发送 SCA 请求。
- `source_kind=directory` 时扫描 `input/`；`source_kind=zip` 时固定扫描 `input/source.zip`。未经协议升级不得依赖其他文件名。
- 请求不得携带任意绝对路径或 `..` 路径。`sca-server` 必须根据 `scan_id` 从固定根目录推导工作目录，并进行规范化、根目录包含和符号链接逃逸检查。
- `SCA_WORK_ROOT` 必须是只供 Core 和 SCA 访问的高敏临时卷，不得暴露为宿主机公共目录、静态文件目录或匿名网络共享。
- 工作目录必须按扫描任务隔离；并发任务不得读取、覆盖或删除彼此的文件。
- 输入发布完成后不得继续原地修改。确需更新时，Core 必须创建新的 attempt 或新的工作目录。
- Core 负责业务层清理和保留策略；SCA 只清理自己产生的临时文件，不得删除其他任务或整个共享根目录。
- 临时卷不是业务事实源。需要跨重启恢复或可追溯保留的源码、原始报告和结果仍按平台数据生命周期写入受控对象存储。

### 2.3 控制接口

默认采用 HTTP/JSON 作为控制协议。HTTP 只传递任务元数据和状态，不承载源码或镜像内容：

```http
POST /v1/scans
Content-Type: application/json
Authorization: Bearer <service-token>

{
  "contract_version": "1.0",
  "scan_id": "uuid",
  "source_kind": "zip",
  "profile": "java-maven",
  "limits": {
    "timeout_seconds": 7200,
    "max_cpu": 4,
    "max_memory_mb": 8192,
    "max_disk_bytes": 42949672960
  }
}
```

`sca-server` 根据 `scan_id` 查找 `<SCA_WORK_ROOT>/scans/<scan_id>`，验证 `READY` 后调用 CLI。`source_kind` 允许值为 `directory`、`zip` 和后续启用的 `oci-archive`。接口至少提供：

- `POST /v1/scans`：幂等提交扫描，返回 `202 Accepted` 和稳定 `scan_id`；
- `GET /v1/scans/{scan_id}`：返回排队、运行、成功、失败或取消状态；
- `POST /v1/scans/{scan_id}/cancel`：终止对应 CLI 子进程并清理临时资源；
- `GET /healthz`：报告服务、CLI 版本和工作目录可访问性。

长扫描不得依赖一个持续数小时的同步 HTTP 连接。`sca-server` 必须支持异步任务状态，Core 通过轮询或受控回调获知完成。CLI 的完整结果写入 `output/`，Core 从共享工作目录读取并向平台数据库归一化，HTTP 响应只返回状态、错误码和结果清单引用。

控制接口必须：

- 仅绑定内部网络，不向普通用户或公网开放；
- 使用服务身份认证，并限制可调用 Core 的服务账号；
- 限制请求体大小、连接数、并发任务数和每任务磁盘配额；
- 对重复 `scan_id`、重复提交、活动任务冲突和过期工作目录返回稳定错误码；
- 记录审计事件，但不记录源码内容、完整路径层级中的敏感项目名或服务令牌。

### 2.4 端到端流程

1. Core 接收源码 ZIP 或 Git 固定 revision，并写入共享卷的 staging 目录。
2. Core 完成大小、哈希、压缩包路径和固定 revision 校验，原子发布为 `scans/<scan_id>/input/`，写入 `READY`。
3. Core 向 `sca-server` 提交只含元数据的 HTTP 请求。
4. `sca-server` 验证服务身份、`scan_id`、工作目录、配额和 CLI 版本，然后启动独立的 `sca` CLI 子进程：

   ```text
   sca analyze \
     --input <SCA_WORK_ROOT>/scans/<scan_id>/input \
     --output <SCA_WORK_ROOT>/scans/<scan_id>/output/result.json
   ```

5. `sca` CLI 从共享卷按流式方式读取输入，执行分析和漏洞匹配，将结果原子写入 `output/`。
6. `sca-server` 监控进程、超时、取消、日志和退出状态，并更新任务状态。
7. Core 读取结果清单并完成标准化、去重、授权校验和数据库落库。
8. Core 按数据生命周期删除共享工作目录；任何成功、失败、超时和取消路径都必须执行清理。

### 2.5 部署边界

共享数据卷方案的首个验收目标是与 Core 同机或同 Pod 放置的 SCA Runner。Compose 必须使用明确命名的数据卷或受控绑定挂载，并保证两个容器使用一致的逻辑根目录。

跨主机 SCA 节点不得通过假设“同名目录自动共享”部署。没有共享块存储、共享文件系统或对象存储下载适配器时，不得宣称支持跨节点扫描。后续扩展必须保持控制接口不变，只替换输入定位适配器。

## 3. 范围

### MVP

- 解析常见清单和锁文件，生成组件坐标与依赖图。
- 支持直接依赖、传递依赖、开发依赖和生产依赖分类。
- 通过内部漏洞快照匹配漏洞，保留源记录 ID 和匹配依据。
- 输出 CycloneDX JSON SBOM、统一 Finding、原始引擎报告。
- 支持固定 Git commit、源码压缩包和可选的容器镜像入口。
- 支持规则/漏洞数据更新、数据版本回滚和离线镜像。

### 后续

- 制品仓库和镜像仓库持续监控。
- Reachability、签名验证、许可证合规和供应链策略。
- 多数据源交叉验证、EPSS/KEV 等风险优先级信号。

### 非目标

- 首版自研所有语言解析器和漏洞匹配算法。
- 自动修改用户依赖或提交 PR。
- 未经许可无限爬取第三方漏洞网站。

## 4. 语言与生态优先级

优先级是建议默认值，待 Q-004 确认：

| 优先级 | 生态 | 输入样例 | MVP 要求 |
|---|---|---|---|
| P0 | Java/Maven/Gradle | `pom.xml`、Gradle lock | 依赖解析、坐标、直接/传递关系 |
| P0 | JS/TS/npm/pnpm/yarn | lockfile、`package.json` | workspace 和 lockfile 语义 |
| P0 | Python/PyPI | `requirements*.txt`、PEP 621、poetry/uv lock | 版本约束和环境标记 |
| P0 | Go Modules | `go.mod`、`go.sum` | module path、replace 和 graph |
| P1 | .NET/NuGet | `.csproj`、`packages.lock.json` | 目标框架和条件依赖 |
| P1 | Rust/Cargo | `Cargo.toml`、`Cargo.lock` | workspace 与 features |
| P1 | PHP/Composer | `composer.json`、`composer.lock` | dev/prod 区分 |
| P1 | Ruby/Bundler | `Gemfile`、`Gemfile.lock` | groups 区分 |
| P2 | 容器镜像 | OCI image | OS 包与语言包、层信息 |
| P2 | 原生包 | RPM/DEB/APK | 仅构建制品模式 |

已确认的首版取舍：只实现 Java/Maven P0。上表中的其他生态是路线图，不代表当前实现。

新增生态必须提供固定样本、已知漏洞回归样本和误报样本。

## 5. 功能需求

### 5.1 输入与预处理

- 平台通过 `scan_id`、固定 revision、扫描策略和共享工作目录定位输入；HTTP 控制请求不传输输入文件。
- 拒绝符号链接逃逸、超大文件、压缩炸弹和路径穿越。
- 默认只读解析，不执行包管理器脚本。
- 有多个清单时保留模块/workspace 所属关系，避免错误合并依赖图。
- 无法解析的文件记录为诊断信息，不得静默当成“无依赖”。

### 5.2 组件识别

每个组件至少包含：

- `ecosystem`、`namespace`、`name`、`version`；
- PURL；可用时包含 CPE，但 CPE 不得替代 PURL；
- 来源文件、依赖路径、直接/传递标记；
- 依赖范围（runtime/dev/build/optional）和解析置信度；
- 许可证声明（可选但保留原始值）。

依赖图应是组件和边组成的有向图；环和重复节点必须有确定的序列化顺序。

### 5.3 SBOM

- 默认输出 CycloneDX JSON，保留工具版本、生成时间和输入 revision。
- 输出可复现：同一输入和同一数据版本产生等价组件集和稳定排序。
- 组件身份尽量使用 PURL；内部包使用明确的 `internal:` 坐标，不伪造公共坐标。
- SBOM 与 Finding 使用同一组件 ID 映射，支持后续增量对比。

### 5.4 漏洞匹配

- 匹配主键优先使用生态 + 包名 + 规范化版本 + 数据源记录 ID。
- 不把 NVD 的 CPE 通用匹配直接当作精确包匹配；必须记录匹配方法和置信度。
- 支持受影响版本范围、修复版本、撤回/替换记录和数据更新时间。
- 一个组件可关联多个漏洞；一个漏洞可影响多个组件。
- 区分 `vulnerable`、`not_affected`、`unknown`，无法判定时不得默认为安全。
- 结果必须携带 CVSS（如源提供）、严重性来源和风险策略版本，避免把不同来源的分值混用而不标注。

### 5.5 去重与证据

- Finding 指纹由生态、包、受影响范围、漏洞源标识和项目上下文构成；不能只用 CVE 标题。
- 同一扫描内相同证据聚合为一个 Finding 和多个 occurrence。
- 保留漏洞源 URL/记录、匹配区间、依赖路径和扫描器原始输出引用。
- 如果数据源撤回漏洞，后续扫描应能把关联 Finding 标记为 `withdrawn` 而不是静默删除历史。

## 6. 漏洞数据与组件信息获取

SCA 引擎不直接在内网任意抓取网页。采用“数据源适配器 -> 校验/规范化 -> 内部快照 -> 扫描器只读使用”的流水线。

### 6.1 数据源分层

1. **包生态元数据**：npm registry、PyPI metadata、Maven Central/Nexus 等，用于组件规范化和许可证/来源信息。
2. **漏洞公告**：OSV、GitHub Advisory、NVD、供应商安全公告、内部漏洞库。
3. **风险增强**：EPSS、CISA KEV、内部资产重要性，可选。
4. **内部豁免/例外**：只存项目级策略，不修改公共漏洞事实。

### 6.2 同步要求

- 每个适配器声明许可证、访问方式、速率限制、增量游标和删除语义。
- 支持 API key 存在于 Secret Manager，不写入配置文件或日志。
- 拉取后先落到隔离区，校验 schema、签名（如有）和版本连续性，再原子切换快照。
- 同步失败不覆盖旧快照；健康检查暴露最后成功时间和数据滞后。
- 记录源记录原始字段的校验和，便于审计和重放。
- 对爬取得到的页面保存 URL、时间、内容哈希和解析器版本；遵守 robots 和服务条款。
- 不依赖单一公网源；数据源下线时扫描仍可用最后一版快照。

### 6.3 数据模型

- `package_coordinate`
- `package_version`
- `vulnerability_source_record`
- `vulnerability`
- `affected_range`
- `fixed_version`
- `alias`
- `data_snapshot`
- `data_sync_run`

所有记录包含 `source`、`source_id`、`fetched_at`、`valid_from/valid_to`、`raw_checksum`。

## 7. 引擎策略

首版通过可替换适配器组合成熟工具，建议候选：

- Syft/Trivy：SBOM、容器和生态覆盖面；
- OSV-Scanner：OSV 坐标匹配；
- Grype：漏洞匹配与数据库；
- 自研归一化层：负责 PURL、内部坐标、策略和稳定 Finding。

候选工具必须经过许可证、离线能力、漏洞数据时效、生态覆盖和误报基准评估后才能入选。不得在同一结果中无标注地重复计数多个工具对同一漏洞的发现。

## 8. 输出契约

每批结果包含：

```json
{
  "schema_version": "1.0",
  "scan_id": "uuid",
  "engine": {"name": "sca", "version": "string"},
  "data_snapshot_id": "string",
  "components": [],
  "findings": [],
  "artifacts": [],
  "diagnostics": []
}
```

Finding 字段和错误模型以平台契约为准。SCA 额外字段见 [../docs/design-docs/scan-result-model.md](../docs/design-docs/scan-result-model.md)。

## 9. 非功能需求

- 500 万行/2 GB 输入在 20 并发下不拖垮平台 API。
- CLI 和 `sca-server` 必须支持 GB 级输入的基于路径处理；禁止把完整 ZIP、镜像归档或扫描源码读入单个 `byte[]`/内存缓冲区。
- 大文件从共享卷到 CLI 的处理必须使用流式读取、临时磁盘或外部引擎原生文件接口；内存占用不得随输入总大小线性增长。
- 输入文件传输不得经过控制 HTTP；控制请求体应保持为小型元数据。磁盘容量、并发和任务时限必须有显式配额与背压。
- 组件识别失败必须可诊断；解析覆盖率作为质量指标。
- 同一输入、同数据快照重复扫描结果稳定，可控地忽略时间戳。
- 默认不联网；需要更新时仅数据同步任务联网。
- 所有临时文件在任务结束后清理，失败和取消路径也清理。
- 漏洞库和 SBOM 不包含平台管理凭据或用户隐私。

## 10. 质量与测试

- **黄金样本**：每种生态至少 5 个真实结构的合成仓库，不提交真实公司代码。
- **CLI 独立测试**：不启动 `sca-server` 和 Core，直接对本地固定样本执行 `sca` 并验证退出码和结果。
- **服务控制测试**：固定样本通过 Core 共享目录和 `sca-server -> CLI` 子进程完成扫描，同时断言 HTTP 请求和响应不包含源码内容。
- **共享卷测试**：覆盖 staged 输入原子发布、缺失/损坏 `READY`、重复 `scan_id`、路径穿越、符号链接逃逸和并发目录隔离。
- **大输入测试**：使用合成的 GB 级 ZIP 或镜像归档验证基于文件的处理路径、磁盘配额、超时、取消和内存上限。
- **漏洞回归**：已知直接依赖、传递依赖、版本边界和撤回公告样本。
- **误报样本**：同名不同生态、fork 包、内部包、不同 major/minor 版本。
- **模糊测试**：清单、lockfile、压缩包路径和超大 JSON/YAML。
- **数据源契约测试**：固定抓取结果回放，验证 schema 和增量游标。
- **可复现性测试**：相同输入和快照结果哈希一致。

建议指标：清单识别率、SBOM 组件准确率、已知样本召回率、漏洞匹配误报率、数据快照滞后、扫描 P95 时长。首版先建立基线，不承诺未经验证的准确率。

## 11. 里程碑

1. （已完成）将 Maven 分析核心抽成可独立运行的 `sca` CLI，保留现有固定样本回归，并停止给 CLI 增加 HTTP 职责。
2. （已完成）实现 `sca-server` 控制接口、CLI 子进程生命周期、幂等提交、状态、取消和服务身份认证。
3. （已完成）将 Core 的源码暂存迁移到共享工作卷，发送只含元数据的控制请求，并从共享结果目录读取结果。
4. 完成 GB 级输入、配额、清理和失败恢复测试。
5. 完成清单解析与 CycloneDX SBOM，接入内部漏洞快照和 PURL 匹配。
6. 输出统一 Finding、依赖路径和 SARIF 映射，并做 50 个样本项目回归。

## 12. 验收标准

- `sca` CLI 可以在没有 Core、`sca-server`、数据库和网络监听的情况下分析本地固定样本。
- 正常扫描中，Core 只向 `sca-server` 发送任务元数据；SCA 通过共享工作目录读取输入，HTTP 日志和请求体中没有源码或镜像内容。
- `sca-server` 通过独立 CLI 子进程完成扫描，不包含重复的分析实现，并能正确处理成功、失败、超时和取消。
- 共享工作目录按 `scan_id` 隔离，缺失就绪标记、重复任务、路径穿越和符号链接逃逸均被拒绝并留有审计记录。
- GB 级输入不会导致服务端将完整文件载入内存，也不会因未经配额限制的并发耗尽磁盘。
- 对每个 P0 生态，黄金样本能识别直接和传递依赖。
- 同一输入重复扫描组件集合稳定，失败能够定位到文件/解析器。
- 已知漏洞样本命中，且结果能追溯数据源、版本范围和修复版本。
- 离线环境能使用内部快照完成扫描。
- 未经授权的网络访问被默认阻断，临时源码和凭据被清理。
