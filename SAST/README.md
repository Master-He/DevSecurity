# SAST

静态应用安全测试（Static Application Security Testing）项目。

阅读 [SPEC.md](SPEC.md)。`joern-adapter/` 已使用本地 Joern 4.0.674 完成 Java 源码基础检测：

- 命令执行：Runtime.exec、ProcessBuilder.start；
- 弱加密：MD5、SHA-1、DES、RC4；
- 不安全反序列化：ObjectInputStream.readObject。

适配器接收 ZIP，执行安全解压，先生成 CPG，再运行仓库内置 CPGQL 规则并返回结构化 Finding。扫描过程不下载 query bundle，也不执行 Maven 依赖获取。

```powershell
..\..\scripts\setup-joern.ps1
go test ./...
go run ./cmd/adapter
```

开发环境默认监听 `127.0.0.1:8087`，服务令牌从 `SAST_SERVICE_TOKEN` 读取。
