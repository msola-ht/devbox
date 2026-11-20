# 开发工具前端索引

展示 Git、uv、Python 和 Homebrew 的检测结果并提供单项刷新操作。

- [dev-tools-page.tsx](dev-tools-page.tsx) 将四张检测卡片排成响应式网格并连接各自的服务调用。
- [tool-card.tsx](tool-card.tsx) 组合工具摘要、刷新按钮、请求提示和详情入口。
- [tool-results.tsx](tool-results.tsx) 渲染检测状态、路径提示与安装详情弹窗。
- [use-tool-detection.ts](use-tool-detection.ts) 管理单项检测的加载、刷新、取消与最近一次结果。

[业务模块索引](../index.md) · [后端模块](../../../internal/devtools/index.md) · [项目架构](../../../ARCHITECTURE.md)
