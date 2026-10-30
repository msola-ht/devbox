# Rust 后端模块说明

[文档索引](../docs/README.md) · [架构设计](../docs/architecture.md) · [后端规则](AGENTS.md) · [前端模块说明](../src/README.md) · [开发指南](../docs/development.md)

本文描述当前 Rust 与 Tauri 实现。当前后端提供 `.devbox` 初始化、环境快照、uv 与 Python 系统检测、托管安装扫描和工具快照缓存；安装、任务、持久化迁移和业务事件按开发指南的实施顺序进入后端模块。

## 启动与组合

```text
main.rs
  devbox_lib::run()
    Tauri Builder
      setup
        homeDir -> ~/.devbox
        ensure_layout
        log -> ~/.devbox/logs
        EnvironmentService
        PluginRegistry(UvPlugin, PythonPlugin) + DevboxLayout + ToolchainService
        manage(AppState)
      invoke_handler
        get_environment / list_tools / refresh_tools
      run
```

| 文件 / 模块 | 当前职责 |
| --- | --- |
| [main.rs](src/main.rs) | 桌面二进制入口 |
| [lib.rs](src/lib.rs) | 唯一组合根：初始化目录、日志、服务、插件、状态和 command |
| [state.rs](src/state.rs) | 保存 `bootId` 和已组装服务；业务缓存归所属服务 |
| [ipc/](src/ipc) | command 边界、响应 DTO 和错误 DTO |
| [modules/environment/](src/modules/environment) | 启动时固定的 OS、架构、用户目录和 `.devbox` 状态 |
| [modules/toolchain/](src/modules/toolchain) | 工具模型、检测应用服务、快照缓存、类型化错误、插件契约和注册表 |
| [infrastructure/path.rs](src/infrastructure/path.rs) | `.devbox` 布局、权限和目录安全校验 |
| [infrastructure/process.rs](src/infrastructure/process.rs) | 受控版本命令执行、超时和输出上限 |

## IPC 契约

| command | 当前行为 |
| --- | --- |
| `get_environment` | 返回当前进程的 `bootId`、固定环境 revision、OS、架构、用户目录、`.devbox` 路径和目录状态 |
| `list_tools` | 调用工具链服务；已有缓存时返回权威快照，无缓存时执行检测并缓存 |
| `refresh_tools` | 调用工具链服务重新检测并分配新 revision |

当前快照 DTO：

- `EnvironmentSnapshot`：`bootId`、`revision`、`generatedAt`、`os`、`arch`、`homeDir`、`devboxRoot`、`devboxReady`、`devboxError`。
- `ToolListSnapshot`：`bootId`、`revision`、`generatedAt`、`tools`。
- `ToolSummary`：`toolId`、`displayName`、`description`、`pluginContractVersion`、系统安装、托管安装和 active 安装 ID。
- `InstallationSummary`：`installationId`、`version`、`path`、`source`、`sourceHint`、`healthy`、`errorCode`。
- `IpcError`：稳定 `code`、可展示 `message` 和可选 `details`。

JSON 字段为 camelCase；标识符、revision 和契约版本以字符串传输；时间使用 UTC RFC 3339。Rust 当前没有自动生成 TypeScript 类型，`src/features/**/contracts.ts` 由人工保持一一对应。

工具链内部模型不派生 Serde；`ipc/toolchain.rs` 显式映射独立 DTO，领域字段和传输字段不共享类型。

当前业务错误码包括 `TOOLS_DIRECTORY_UNAVAILABLE`、`TOOLS_DIRECTORY_UNSAFE` 和用于阻塞任务异常的 `TOOLS_SCAN_FAILED`；安装项诊断包括 `SYSTEM_PROBE_FAILED`、`VERSION_OUTPUT_UNRECOGNIZED` 和 `INSTALLATION_DAMAGED`。前端传输层另有 `IPC_TRANSPORT_FAILED`、`DESKTOP_UNAVAILABLE` 和 `STALE_SNAPSHOT`。

## 工具插件

`ToolPlugin` 当前契约：

| 能力 | 说明 |
| --- | --- |
| `descriptor` | 稳定 `toolId`、显示名、说明和契约版本 |
| `executableNames` | 平台相关可执行文件名 |
| `knownDirectories` | 插件声明的常见绝对目录 |
| `versionArgs` | 版本探测参数 |
| `parseVersionOutput` | 工具专有输出解析 |
| `sourceHint` | 根据规范路径生成来源提示 |

`lib.rs` 显式创建 `PluginRegistry` 并注册 `UvPlugin`、`PythonPlugin`；插件接收 OS、用户目录等最小探测上下文，而不依赖完整环境模型。通用服务负责 PATH 合并、路径规范化与去重、只读进程探测、清单读取和领域摘要，IPC 边界再映射响应 DTO。服务不包含 uv 或 Python 的输出解析规则。

Python 提供系统解释器检测。版本目录、安装、验证和激活通过后续插件能力与 `operations` 服务接入；`operations` 只使用 `toolchain` 公开的能力接口。

## 用户目录与状态

`ensure_layout` 创建：

```text
~/.devbox/
  bin/ tools/ state/{active,operations,locks}/
  cache/{releases,downloads}/ staging/ backups/ logs/
```

- Unix 下根目录和全部子目录均为 `0700`。
- 已存在的符号链接目录会被拒绝，避免初始化写入逃逸。
- 托管工具根目录同样拒绝符号链接。
- 托管清单必须有 `schemaVersion`、匹配的 `toolId` 和 `pluginContractVersion`、安装 ID、版本、可执行相对路径和健康标记。
- 可执行路径必须是纯相对路径，规范化后仍位于对应安装目录内。
- active 指针读取 `state/active/<tool>.json`，校验格式版本、工具归属，并只在指向健康托管安装时对外返回。
- `DevboxLayout` 集中提供托管工具根目录和 active 指针位置，工具链服务不直接拼接目录语义。

当前没有写入 manifest、active 指针或 `.devbox/bin` 的代码，也没有原子替换、备份迁移或跨进程锁。

安装链路接入后，托管安装清单与 active 指针的唯一写入入口归 `toolchain` 库存服务；`operations` 提交验证通过的安装计划，不直接写这些文件。

## 并发与缓存

- `bootId` 在每次 Rust 进程启动时生成。
- 工具链 revision 和最新快照缓存归 `ToolchainService`，revision 在检测开始时分配并在进程内单调递增。
- 并发刷新不会让较低 revision 覆盖已缓存的较高 revision。
- 系统安装候选发现和托管安装扫描通过 `spawn_blocking` 脱离异步执行器，避免阻塞文件操作占住 Tokio 工作线程。
- 前端快照 Store 合并同一业务查询；Rust 单飞探测、任务注册表和跨进程锁随操作服务建立。
- 前端断连或被销毁不影响当前只读检测；后台写任务归 `operations` 状态机。

## 日志与安全边界

- 日志插件在 `.devbox` 初始化成功后启用，目标为 `.devbox/logs/devbox.log`，单文件上限 10 MiB。
- 版本命令使用绝对可执行路径和参数数组，不使用 Shell；清空环境后仅保留 Windows 运行所需的最小系统变量。
- 版本探测有 3 秒超时、stdout/stderr 各 64 KiB 上限，并终止直接子进程。
- PATH 中的相对目录会被跳过，避免执行当前工作目录中的同名程序。
- `build.rs` 为三个自定义 command 生成 allow/deny ACL；`capabilities/default.json` 只向 `main` 窗口放行这三个命令。

## 构建与配置

| 文件 | 当前配置 |
| --- | --- |
| [Cargo.toml](Cargo.toml) | edition 2024、Tauri 2、Serde、Tokio、UUID、time 和日志依赖 |
| [build.rs](build.rs) | Tauri build 和自定义 command AppManifest |
| [tauri.conf.json](tauri.conf.json) | DevBox 标识、主窗口、开发地址、前端产物和打包配置 |
| [capabilities/default.json](capabilities/default.json) | main 窗口的 `core:default` 与三个业务命令权限 |
| `tauri.<platform>.conf.json` | 各平台打包目标；没有实际原生验证前不代表已支持 |

当前 CSP 仍为 `null`。没有远程内容，但正式发布前仍需配置并验证收紧后的 CSP。

## 后续交付

- Release 目录、资产选择、下载、摘要校验、解压、安装、更新和取消进入工具安装链路。
- 操作任务、channel、event、重连查询、启动恢复和跨进程锁进入 `operations` 聚合。
- 持久化格式迁移、备份恢复、卸载、缓存清理和修复命令进入完整工具管理能力。
- Python 复用 uv 托管运行时；Node.js 和 Git 作为独立插件接入。
- Windows 和 Linux 在目标平台完成原生构建与运行验证。
- TypeScript IPC 类型和运行时契约校验由契约生成流程维护。
