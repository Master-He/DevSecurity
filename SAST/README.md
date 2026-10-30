# SAST

静态应用安全测试（Static Application Security Testing）项目。

阅读 [SPEC.md](SPEC.md)。首批引擎确定为 Joern，`joern-adapter/` 提供对官方 CLI 的隔离接入骨架。该适配器当前只完成 Java 源码到 Code Property Graph 的构建和健康检查，漏洞规则查询与统一 Finding 映射仍是下一阶段工作；未安装 Joern 时会明确返回不可用，不会伪造“未发现漏洞”。
