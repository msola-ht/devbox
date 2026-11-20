# 前端索引

React 界面入口、路由和全局样式。业务通过 MyGo 生成接口调用 Go 服务。

- [main.tsx](main.tsx)：恢复路由，挂载 React 和 HashRouter。
- [app.tsx](app.tsx)：组合应用外壳、导航和页面。
- [app-routes.ts](app-routes.ts)：路由定义、路径校验和页面记忆。
- [mygo.ts](mygo.ts)：由 Go 定义生成的服务接口和数据类型。
- [style.css](style.css)：全局样式、主题变量和字体。
- [features](features/index.md)：按业务组织的页面、组件和 Hook。
- [components](components/index.md)：通用界面组件。
- [hooks](hooks/index.md)：跨业务复用的 React Hook。
- [lib](lib/index.md)：通用前端工具。

[项目架构](../ARCHITECTURE.md) · [后端索引](../internal/index.md)
