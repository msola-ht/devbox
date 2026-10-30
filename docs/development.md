# DevBox 开发指南

[文档索引](README.md) · [前端模块说明](../src/README.md) · [后端模块说明](../src-tauri/README.md)

本文记录开发、构建、组件生成与验证操作。下列命令均在仓库根目录执行。

## 架构约束

产品结构和模块边界见[架构设计](architecture.md)。架构与实现约束分别定义在 [根目录规则](../AGENTS.md)、[前端规则](../src/AGENTS.md) 和 [Rust 后端规则](../src-tauri/AGENTS.md)。根规则适用于整个仓库，子目录规则补充对应范围；涉及前后端通信的变更同时受两侧约束。

规则中的目录树定义后续扩展的职责边界，不表示业务模块、IPC 或状态系统已经实现。当前开发与验证操作说明集中在本文件。

## 开发

```bash
bun install
bun tauri dev
```

仅调试前端：

```bash
bun run dev
```

## 多平台构建

Tauri 会按当前操作系统自动加载 `src-tauri/tauri.<platform>.conf.json`。应在目标系统上执行原生构建，不建议从单一系统直接交叉打包桌面安装文件。

| 平台 | 命令 | 产物 |
| --- | --- | --- |
| macOS | `bun tauri build` | `.app`、`.dmg` |
| Linux | `bun tauri build` | AppImage、`.deb`、`.rpm` |
| Windows | `bun tauri build` | NSIS `.exe`、MSI |

macOS 通用二进制需要先安装两种架构的 Rust target：

```bash
rustup target add aarch64-apple-darwin x86_64-apple-darwin
bun run tauri:build:macos:universal
```

Ubuntu 22.04 或更新版本构建前可安装 Tauri 系统依赖：

```bash
sudo apt update
sudo apt install libwebkit2gtk-4.1-dev build-essential curl wget file libxdo-dev libssl-dev libayatana-appindicator3-dev librsvg2-dev
```

安装包默认未配置 macOS 公证和 Windows 代码签名。发布前需在对应构建环境中补充证书与签名密钥。

## 添加 UI 组件

使用 shadcn CLI 添加组件：

```bash
bunx --bun shadcn@latest add button
```

组件会生成到 `src/components/ui`。

可能覆盖已有组件或 block 时，先预览差异，再应用覆盖：

```bash
bunx --bun shadcn@latest add <component-or-block> --dry-run
bunx --bun shadcn@latest add <component-or-block> --overwrite
```

## 开发验证

UI 或路由修改后，至少执行前端构建：

```bash
bun run build
```

使用当前环境可用的浏览器工具或技能打开目标路由，检查全局布局继承、主题、操作结果及视口滚动。使用技能前读取其 `SKILL.md`。没有可用浏览器能力时，记录未验证项和原因，不能以构建成功代替页面观察。

涉及 IPC 或 Rust 业务行为时，通过 `bun tauri dev` 观察真实桌面操作从入口到持久化或任务结果；普通浏览器页面只能覆盖前端表现，不能证明 Rust 通信成功。根据实际变更选择已有静态检查，并分别记录静态结果与行为观察；不新增或修改测试，不以测试结果作为完成依据。

纯规则与文档修改检查内容一致性、引用路径和适用范围，无需运行无关的应用流程。

通信机制参考 Tauri 官方文档：[前端调用 Rust](https://v2.tauri.app/develop/calling-rust/)、[Rust 调用前端](https://v2.tauri.app/develop/calling-frontend/)和[能力权限](https://v2.tauri.app/security/capabilities/)。
