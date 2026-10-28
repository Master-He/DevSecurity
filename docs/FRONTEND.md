# 前端工程说明

完整产品行为见 [platform/frontend/SPEC.md](../platform/frontend/SPEC.md)，视觉原则见 [DESIGN.md](DESIGN.md)。

## 技术基线

- React + TypeScript + Vite；
- TanStack Query 管理服务端状态，TanStack Router 管理路由；
- Ant Design 提供企业后台基础组件；
- ECharts 用于趋势和分布；
- Vitest + Testing Library 做单元/组件测试；
- Playwright 做关键路径端到端测试。

## 代码组织建议

```text
src/
  app/            路由、启动、全局 provider
  features/       按业务能力组织的页面、查询和状态
  components/     跨业务复用组件
  api/            OpenAPI 生成的 client 和查询封装
  styles/         theme、token、全局样式
  i18n/           文案和格式
  test/           测试工具和固定数据
```

`features` 不能反向导入页面私有实现。服务端字段由生成的 client 约束，不能手写多份不一致的接口类型。

## 状态边界

- 扫描结果、权限、项目信息属于服务端状态，交给 Query。
- URL 保存可分享的筛选、排序和分页游标。
- 表单临时状态保留在组件/表单层。
- 全局只保存主题、语言、导航折叠等非敏感偏好。

## 安全要求

- 不把访问令牌或 Cookie 写入 localStorage。
- 不把源码、证据或 Secret 写入前端日志和第三方埋点。
- 后端返回 HTML/Markdown 时必须净化，不直接 `dangerouslySetInnerHTML`。
- 所有下载由后端授权并设置短过期。
- 前端权限只是 UX 辅助，后端必须再次校验。

## 质量门禁

- TypeScript 严格模式、lint 和格式检查无错误；
- 所有数据请求能处理超时、401/403、429 和 5xx；
- 关键页面具有无障碍名称、键盘操作和可见焦点；
- 核心场景具备 E2E 测试；
- 构建可离线部署，不依赖公网 CDN 和字体服务。
