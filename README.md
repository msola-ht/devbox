# Devbox

Devbox 是一个面向 macOS、Windows 和 Linux 的桌面开发环境工具箱，用于查看本机系统信息，并逐步支持检测和安装 Git、Python、uv 等开发工具。

目前已实现系统信息采集、展示、本地快照和应用级代理设置；工具检测与安装尚在设计阶段。项目使用 MyGo、Go 和 React 构建。

## 开发

准备 Go（版本见 [go.mod](go.mod)）、Bun 和 MyGo 对应平台的构建依赖，然后执行：

```sh
bun install
bun run dev
```

[产品设计](DESIGN.md) · [当前架构](ARCHITECTURE.md) · [项目规则](AGENTS.md)
