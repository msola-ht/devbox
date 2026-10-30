# DevBox

DevBox 是多平台开发工具链管理器，目标能力包括系统检测、托管安装、版本切换、更新和回滚。默认把托管工具、缓存、状态和日志保存在用户目录的 `.devbox`。

当前 MVP 第一个切片已完成 macOS 环境与 uv 只读检测，并提供概览和工具页面。下载、安装、更新、任务进度和多平台原生验证尚未实现。

技术栈：React、TypeScript、Vite、shadcn/ui、Bun、Tauri 2 与 Rust。

## 文档

- [架构设计](docs/architecture.md)：平台、领域模型、工具插件、用户目录、持久化、通信同步、安全、发布和 MVP 全链路。
- [文档索引](docs/README.md)：开发指南、前后端模块说明与架构规则的统一入口。
- [环境检测实施归档](docs/archive/2026-10-29-mvp-01-environment-detection.md)：当前能力、审查修复、验证证据和剩余限制。
