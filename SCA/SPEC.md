# SCA 引擎 SPEC

## 1. 目的

从源代码、锁文件、构建产物和容器镜像中识别第三方组件及其依赖关系，匹配已知漏洞，输出可去重、可追溯、可采取行动的风险结果和 SBOM。

SCA 不能仅凭“存在某版本库”就断言业务可利用；结果应携带证据、置信度、直接/传递依赖关系和修复建议。

## 2. 范围

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

## 3. 语言与生态优先级

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

新增生态必须提供固定样本、已知漏洞回归样本和误报样本。

## 4. 功能需求

### 4.1 输入与预处理

- 由平台传入固定 revision、项目 ID、扫描策略和对象存储输入位置。
- 拒绝符号链接逃逸、超大文件、压缩炸弹和路径穿越。
- 默认只读解析，不执行包管理器脚本。
- 有多个清单时保留模块/workspace 所属关系，避免错误合并依赖图。
- 无法解析的文件记录为诊断信息，不得静默当成“无依赖”。

### 4.2 组件识别

每个组件至少包含：

- `ecosystem`、`namespace`、`name`、`version`；
- PURL；可用时包含 CPE，但 CPE 不得替代 PURL；
- 来源文件、依赖路径、直接/传递标记；
- 依赖范围（runtime/dev/build/optional）和解析置信度；
- 许可证声明（可选但保留原始值）。

依赖图应是组件和边组成的有向图；环和重复节点必须有确定的序列化顺序。

### 4.3 SBOM

- 默认输出 CycloneDX JSON，保留工具版本、生成时间和输入 revision。
- 输出可复现：同一输入和同一数据版本产生等价组件集和稳定排序。
- 组件身份尽量使用 PURL；内部包使用明确的 `internal:` 坐标，不伪造公共坐标。
- SBOM 与 Finding 使用同一组件 ID 映射，支持后续增量对比。

### 4.4 漏洞匹配

- 匹配主键优先使用生态 + 包名 + 规范化版本 + 数据源记录 ID。
- 不把 NVD 的 CPE 通用匹配直接当作精确包匹配；必须记录匹配方法和置信度。
- 支持受影响版本范围、修复版本、撤回/替换记录和数据更新时间。
- 一个组件可关联多个漏洞；一个漏洞可影响多个组件。
- 区分 `vulnerable`、`not_affected`、`unknown`，无法判定时不得默认为安全。
- 结果必须携带 CVSS（如源提供）、严重性来源和风险策略版本，避免把不同来源的分值混用而不标注。

### 4.5 去重与证据

- Finding 指纹由生态、包、受影响范围、漏洞源标识和项目上下文构成；不能只用 CVE 标题。
- 同一扫描内相同证据聚合为一个 Finding 和多个 occurrence。
- 保留漏洞源 URL/记录、匹配区间、依赖路径和扫描器原始输出引用。
- 如果数据源撤回漏洞，后续扫描应能把关联 Finding 标记为 `withdrawn` 而不是静默删除历史。

## 5. 漏洞数据与组件信息获取

SCA 引擎不直接在内网任意抓取网页。采用“数据源适配器 -> 校验/规范化 -> 内部快照 -> 扫描器只读使用”的流水线。

### 5.1 数据源分层

1. **包生态元数据**：npm registry、PyPI metadata、Maven Central/Nexus 等，用于组件规范化和许可证/来源信息。
2. **漏洞公告**：OSV、GitHub Advisory、NVD、供应商安全公告、内部漏洞库。
3. **风险增强**：EPSS、CISA KEV、内部资产重要性，可选。
4. **内部豁免/例外**：只存项目级策略，不修改公共漏洞事实。

### 5.2 同步要求

- 每个适配器声明许可证、访问方式、速率限制、增量游标和删除语义。
- 支持 API key 存在于 Secret Manager，不写入配置文件或日志。
- 拉取后先落到隔离区，校验 schema、签名（如有）和版本连续性，再原子切换快照。
- 同步失败不覆盖旧快照；健康检查暴露最后成功时间和数据滞后。
- 记录源记录原始字段的校验和，便于审计和重放。
- 对爬取得到的页面保存 URL、时间、内容哈希和解析器版本；遵守 robots 和服务条款。
- 不依赖单一公网源；数据源下线时扫描仍可用最后一版快照。

### 5.3 数据模型

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

## 6. 引擎策略

首版通过可替换适配器组合成熟工具，建议候选：

- Syft/Trivy：SBOM、容器和生态覆盖面；
- OSV-Scanner：OSV 坐标匹配；
- Grype：漏洞匹配与数据库；
- 自研归一化层：负责 PURL、内部坐标、策略和稳定 Finding。

候选工具必须经过许可证、离线能力、漏洞数据时效、生态覆盖和误报基准评估后才能入选。不得在同一结果中无标注地重复计数多个工具对同一漏洞的发现。

## 7. 输出契约

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

## 8. 非功能需求

- 500 万行/2 GB 输入在 20 并发下不拖垮平台 API。
- 组件识别失败必须可诊断；解析覆盖率作为质量指标。
- 同一输入、同数据快照重复扫描结果稳定，可控地忽略时间戳。
- 默认不联网；需要更新时仅数据同步任务联网。
- 所有临时文件在任务结束后清理，失败和取消路径也清理。
- 漏洞库和 SBOM 不包含平台管理凭据或用户隐私。

## 9. 质量与测试

- **黄金样本**：每种生态至少 5 个真实结构的合成仓库，不提交真实公司代码。
- **漏洞回归**：已知直接依赖、传递依赖、版本边界和撤回公告样本。
- **误报样本**：同名不同生态、fork 包、内部包、不同 major/minor 版本。
- **模糊测试**：清单、lockfile、压缩包路径和超大 JSON/YAML。
- **数据源契约测试**：固定抓取结果回放，验证 schema 和增量游标。
- **可复现性测试**：相同输入和快照结果哈希一致。

建议指标：清单识别率、SBOM 组件准确率、已知样本召回率、漏洞匹配误报率、数据快照滞后、扫描 P95 时长。首版先建立基线，不承诺未经验证的准确率。

## 10. 里程碑

1. 确定首批生态、许可证和数据源，建立黄金仓库。
2. 完成清单解析与 CycloneDX SBOM。
3. 接入内部漏洞快照和 PURL 匹配。
4. 输出统一 Finding、依赖路径和 SARIF 映射。
5. 做 50 个样本项目回归，决定是否替换/增加候选引擎。

## 11. 验收标准

- 对每个 P0 生态，黄金样本能识别直接和传递依赖。
- 同一输入重复扫描组件集合稳定，失败能够定位到文件/解析器。
- 已知漏洞样本命中，且结果能追溯数据源、版本范围和修复版本。
- 离线环境能使用内部快照完成扫描。
- 未经授权的网络访问被默认阻断，临时源码和凭据被清理。
