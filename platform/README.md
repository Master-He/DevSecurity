# Platform

DevSecurity 的控制面，由 Spring Cloud 后端和 React Web 前端组成。当前已包含可运行 PoC。

- [backend/SPEC.md](backend/SPEC.md)
- [frontend/SPEC.md](frontend/SPEC.md)
- [../docs/product-specs/index.md](../docs/product-specs/index.md)

运行完整栈使用仓库根目录的 `scripts/dev-up.ps1` 或 `scripts/dev-up.sh`。

平台是任务与结论的控制面，不直接执行不可信代码，不直接攻击业务目标；下载扫描产物必须经过后端授权。
