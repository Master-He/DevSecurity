# 0002 平台与 Maven SCA PoC

状态：进行中  
开始日期：2026-10-29

## 已完成

- [x] Spring Cloud 多模块结构：Eureka、Gateway、Core
- [x] 本地账号、JWT 和项目级访问控制
- [x] 项目创建和扫描任务列表
- [x] ZIP 上传安全校验
- [x] Git HTTP(S) 地址白名单接入
- [x] PostgreSQL 条件更新任务抢占
- [x] MinIO 源码存储和本地测试存储适配
- [x] Go Maven SCA：dependency tree 优先、POM 回退
- [x] 缺失版本输出 `unknown`
- [x] 离线 PURL 漏洞匹配
- [x] 组件、漏洞、受影响版本、修复版本和依赖路径 API
- [x] 源码 3 天、结论 180 天生命周期字段和清理任务
- [x] React 登录、项目、扫描、组件和 Finding 页面
- [x] Joern CPG 接入骨架
- [x] 后端、前端和 Go 构建/测试
- [x] 本地 H2/文件存储 API 闭环验证
- [x] 本地 HTTP Git 拉取与异步 SCA 成功路径验证
- [x] Maven dependency tree 成功路径及传递依赖验证
- [x] 浏览器登录、建项目、上传和结果页面验证

## 待完成

- [ ] 在可用 Docker Engine 上执行 Compose 双网关/双核心验证
- [ ] 使用真实 PostgreSQL 和 MinIO 进行集成测试
- [ ] 双实例并发领取同一任务的压力/竞争测试
- [ ] 将上述浏览器流程固化为 Playwright 自动化测试
- [ ] 对 3 天/180 天清理执行可控时钟测试
- [ ] 加入 OpenAPI 文档和 SARIF/JSON 导出
- [ ] 拆分并补齐审计事件写入
- [ ] Joern 版本固定、样例基准和规则查询
- [ ] 生产级数据库/对象存储高可用方案

## 验收证据

- Maven：`platform/backend` 全部模块编译成功；
- React：`tsc --noEmit` 与 `vite build` 成功；
- Go：`go test ./...`、`go vet ./...`、`go build` 成功；
- API：登录 -> 项目 -> ZIP 上传 -> 异步 SCA -> 组件/Finding 查询成功；
- Git：登录 -> 项目 -> 本地 HTTP Git 拉取 -> 异步 SCA -> 5 个组件成功；
- 样例：19 个 Maven 组件、5 个已知漏洞、Maven 失败回退 POM；
- Maven：3 个直接依赖、2 个传递依赖，保留依赖路径；
- UI：登录、项目、上传、完成状态与表格均通过浏览器检查；
- 限制：交付时 Docker Engine 未成功启动，Compose 启动仍需环境验证。
