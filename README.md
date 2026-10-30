# DevBox

DevBox 是多平台开发工具链管理器，目标能力包括系统检测、托管安装、版本切换、更新和回滚。默认把托管工具、缓存、状态和日志保存在用户目录的 `.devbox`。

当前版本提供 macOS 环境与默认 Shell 登录环境刷新、uv、Python、Node.js、npm、pnpm、Git、Bash 与 Zsh 只读检测、工具链展示和下载代理设置。下载、安装、任务和多平台原生验证按开发指南的实施顺序推进。

技术栈：React、TypeScript、Vite、shadcn/ui、Bun、Tauri 2 与 Rust。

## 文档

- [架构设计](docs/architecture.md)：DevBox 唯一架构基线，覆盖平台、领域模型、工具插件、用户目录、持久化、通信、安全与发布。
- [开发指南](docs/development.md)：当前开发基线、模块接入方式、后续实施顺序和验证流程。
- [文档索引](docs/README.md)：开发指南、前后端模块说明与架构规则的统一入口。
