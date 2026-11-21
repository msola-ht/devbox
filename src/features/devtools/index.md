# 开发工具前端索引

定义工具目录，在开发工具页展示检测结果，提供单项刷新和 macOS 上的 uv 安装操作。

- [dev-tools-page.tsx](dev-tools-page.tsx) 切换工具作用域，按概览确认的工作目录展示项目工具。
- [tool-catalog.ts](tool-catalog.ts) 定义工具分组、顺序、检测方法和平台适用范围。
- [tool-card.tsx](tool-card.tsx) 组合工具摘要、单项刷新、请求提示和详情入口。
- [tool-results.tsx](tool-results.tsx) 渲染检测状态、安装来源、路径提示和安装详情弹窗。
- [uv-install-action.tsx](uv-install-action.tsx) 在 uv 卡片展示安装入口、任务状态、取消和输出。
- [use-uv-install.ts](use-uv-install.ts) 启动或取消任务，读取任务状态，并在结束后更新卡片的检测结果。

[业务模块索引](../index.md) · [后端模块](../../../internal/devtools/index.md) · [项目架构](../../../ARCHITECTURE.md)
