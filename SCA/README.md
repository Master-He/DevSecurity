# SCA

软件成分分析（Software Composition Analysis）项目。已实现 Maven-only Go 分析核心、独立 CLI、HTTP 控制服务和离线漏洞数据 PoC。

阅读 [SPEC.md](SPEC.md)。实现前需先确认首批生态、漏洞数据源许可和离线更新要求。

项目构建两个可执行文件。`sca`/`sca.exe` 是可直接处理本地目录和 ZIP 的 CLI；`sca-server`/`sca-server.exe` 提供内部控制接口并调用 CLI。Core 与 SCA 通过按 `scan_id` 隔离的共享临时工作卷交换大文件，HTTP 不承载源码或镜像内容。

CLI 用法：

```powershell
.\sca.exe analyze --input ..\examples\java-vulnerable --output .\result.json
```

```bash
go test ./...
go run ./cmd/sca analyze --input ../examples/java-vulnerable --output ./result.json
```

控制服务本地默认监听 `127.0.0.1:8086`：

```powershell
$env:SCA_SERVICE_TOKEN = "local-only-change-me"
$env:SCA_WORK_ROOT = "D:\Project\dev-security\.data\sca-work"
$env:SCA_CLI_PATH = ".\sca.exe"
.\sca-server.exe
```

控制接口只接受任务元数据，默认工作目录结构为：

```text
<SCA_WORK_ROOT>/scans/<scan_id>/
├── input/
│   └── source.zip       # source_kind=zip 时使用
├── output/
├── logs/
└── READY
```

`source_kind=directory` 时扫描 `input/` 目录，`source_kind=zip` 时扫描 `input/source.zip`。Core 必须完成输入发布并写入 `READY` 后再提交任务。漏洞数据文件默认是 `data/vulnerabilities.json`。Core 已接入新控制接口；本地 IDEA 联调时，Core 和 `sca-server` 必须配置完全相同的 `SCA_WORK_ROOT` 绝对路径和 `SCA_SERVICE_TOKEN`。
