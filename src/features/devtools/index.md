# 开发工具前端索引

定义工具目录，在开发工具页展示检测结果，提供单项刷新和按平台适用的本机安装操作。

- [dev-tools-page.tsx](dev-tools-page.tsx) 按平台分组展示工具，读取适用安装渠道，并协调 Node.js 和 npm 的关联刷新。
- [tool-catalog.ts](tool-catalog.ts) 定义工具分组、顺序、检测方法和展示平台。
- [tool-card.tsx](tool-card.tsx) 组合工具摘要、单项刷新、请求提示和详情入口。
- [tool-results.tsx](tool-results.tsx) 渲染检测状态、安装来源、路径提示和安装详情弹窗。
- [tool-install-action.tsx](tool-install-action.tsx) 展示当前渠道的安装入口、状态、取消和输出。
- [use-tool-install.ts](use-tool-install.ts) 提交或取消任务，切页后恢复任务读取，并在结束后更新卡片检测结果。

[业务模块索引](../index.md) · [后端模块](../../../internal/devtools/index.md) · [项目架构](../../../ARCHITECTURE.md)
