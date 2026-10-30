# 开发日志：Python 检测插件与工具链基线

[文档索引](../README.md) · [架构设计](../architecture.md) · [开发指南](../development.md)

本文是历史开发日志，记录当时的实施、审查和验证证据；当前架构决策以[架构设计](../architecture.md)为唯一基线。

状态：已归档

日期：2026-10-29

## 目标

以 Python 作为第二个只读检测工具，验证通用工具链、插件注册、IPC DTO 和前端页面之间的扩展边界，并形成后续安装能力的实施基线。

## 架构审查

- 工具检测链路保持“页面 → app 组合层 → feature → API 适配 → IPC → ToolchainService → 插件”的依赖方向。
- Python 通过插件接入统一 `list_tools` 快照，前端组件、快照 store 和 IPC DTO 保持通用。
- 安装、更新、切换、取消和恢复归独立的 `operations` 聚合，`ToolchainService` 继续负责查询与检测。
- 工具能力按描述、检测、Release、安装计划、验证和激活拆分，插件返回声明式计划，公共执行器负责副作用。
- Rust DTO 作为契约源，TypeScript 类型通过生成流程维护。
- 工具链缓存、revision 和检测日志归 `ToolchainService`；IPC 负责 DTO、错误码和边界转换。
- 阻塞文件扫描通过 `spawn_blocking` 执行，路径语义集中由 `DevboxLayout` 提供。

## 实施记录

Python 插件提供以下检测能力：

- 稳定工具 ID `python`、显示名称和插件契约版本。
- macOS、Linux 和 Windows 的可执行文件名与常见目录。
- Homebrew、Python Framework、用户目录、pyenv、Linux 系统目录和 PATH 的探测。
- `Python 3.x.y` 版本输出解析。
- pyenv、uv 管理 Python、Homebrew、系统 Python、虚拟环境、用户安装和 PATH 来源提示。

插件在 `lib.rs` 与 uv 插件共同注册，工具链服务按统一流程完成 PATH 合并、规范化、去重、版本探测、托管清单扫描和快照构造。

## 验证

静态检查与构建：

- `cargo check`：通过。
- `cargo clippy --all-targets -- -D warnings`：通过。
- `bun run build`：通过。
- `bun run lint`：通过。
- `bun tauri build --debug --no-bundle`：通过。

运行期工具快照：

```text
工具检测完成 revision=1:
  uv (1 system, 0 managed),
  python (2 system, 0 managed)
```

Python 当前提供系统解释器检测。托管清单扫描使用通用服务，安装、激活、切换和卸载由后续 `operations` 能力承担。

## 后续

开发顺序以[开发指南](../development.md)为准：

1. 拆分工具生命周期能力并生成 TypeScript IPC 契约。
2. 建立 `operations` 域、任务状态机和 dry-run 安装计划。
3. 完成 uv 的下载、校验、解压、验证、激活和恢复链路。
4. Python 复用 uv 托管运行时，并接入安装与激活计划。
5. Node.js、Git 和更多查询资源按独立业务切片接入。
