# Rust 后端模块说明

[文档索引](../docs/README.md) · [架构设计](../docs/architecture.md) · [后端规则](AGENTS.md) · [前端模块说明](../src/README.md) · [开发指南](../docs/development.md)

本文描述当前 Rust 与 Tauri 实现。当前后端提供 `.devbox` 初始化、环境快照、uv、Python、Node.js、npm、pnpm、Git、Bash 与 Zsh 系统检测、托管安装扫描、工具快照缓存、代理设置持久化和连接测试；安装、任务、完整持久化迁移和业务事件按开发指南的实施顺序进入后端模块。

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
        SettingsService -> config.json
        PluginRegistry(8 个工具插件) + DevboxLayout + ToolchainService
        manage(AppState)
      invoke_handler
        get_environment / get_settings / update_settings / test_proxy
        / list_tools / refresh_tools
      run
```

| 文件 / 模块 | 当前职责 |
| --- | --- |
| [main.rs](src/main.rs) | 桌面二进制入口 |
| [lib.rs](src/lib.rs) | 唯一组合根：初始化目录、日志、服务、插件、状态和 command |
| [state.rs](src/state.rs) | 保存 `bootId` 和已组装服务；业务缓存归所属服务 |
| [ipc/](src/ipc) | command 边界、响应 DTO 和错误 DTO |
| [modules/environment/](src/modules/environment) | 启动时固定的 OS、硬件规格、用户目录和 `.devbox` 状态 |
| [modules/settings/](src/modules/settings) | 代理配置模型、校验、版本化持久化、原子写入和连接测试 |
| [modules/toolchain/](src/modules/toolchain) | 工具模型、检测应用服务、快照缓存、类型化错误、插件契约和注册表 |
| [infrastructure/path.rs](src/infrastructure/path.rs) | `.devbox` 布局、权限和目录安全校验 |
| [infrastructure/process.rs](src/infrastructure/process.rs) | 受控版本命令执行、超时和输出上限 |
| [infrastructure/network/](src/infrastructure/network) | 读取系统与环境代理，并根据系统、直连或手动代理路由构建 HTTPS 客户端 |

系统规格在应用启动时检测一次；当前 macOS 覆盖版本、内存、卷容量、GPU 和账户默认 Shell，Linux 覆盖版本、内存、容量和账户默认 Shell，Windows 仅覆盖版本。默认 Shell 在 macOS 从用户账户记录读取，不依赖启动应用的交互 Shell 环境。Rust 随后通过默认 Bash/Zsh 的登录环境读取 PATH，用于解析每个工具当前生效的系统安装。未实现的字段以 `null` 返回。

## IPC 契约

| command | 当前行为 |
| --- | --- |
| `get_environment` | 返回启动时检测的 `bootId`、OS 版本、架构、内存、主目录所在卷容量、GPU、用户目录、`.devbox` 路径和目录状态 |
| `get_settings` | 返回代理配置、revision 和当前生效的系统代理读取结果 |
| `update_settings` | 校验、持久化代理配置并返回包含系统代理读取结果的新 revision |
| `test_proxy` | 使用当前输入访问固定 GitHub HTTPS 目标，不读取或修改已保存配置 |
| `list_tools` | 调用工具链服务；已有缓存时返回权威快照，无缓存时执行检测并缓存 |
| `refresh_tools` | 调用工具链服务重新检测并分配新 revision |

当前快照 DTO：

- `EnvironmentSnapshot`：`bootId`、`revision`、`generatedAt`、`os`、`osVersion`、`osBuild`、`arch`、`homeDir`、`devboxRoot`、`devboxReady`、`devboxError`、`memoryTotalBytes`、`diskTotalBytes`、`gpuName`、`gpuCoreCount`、`defaultShell`。
- `ToolListSnapshot`：`bootId`、`revision`、`generatedAt`、`tools`。
- `ToolSummary`：`toolId`、`displayName`、`description`、`pluginContractVersion`、系统安装、托管安装、active 安装 ID 和 `currentInstallationId`。
- `InstallationSummary`：`installationId`、`version`、`path`、`source`、`sourceHint`、`healthy`、`errorCode`。
- `SettingsSnapshot`：`bootId`、`revision`、`generatedAt`、`proxy` 和 `systemProxy`。
- `ProxySettings`：`system`、`direct`，或带 `protocol`、`host`、`port`、`bypass` 的 `manual`。
- `SystemProxyInspection`：`checkedAt`、HTTP/HTTPS 生效代理；每项包含脱敏 URL、`environment` 或 `operatingSystem` 来源和 `requiresAuthentication`。
- `ProxyTestResult`：`checkedAt`、`latencyMs` 和 `httpStatus`。
- `IpcError`：稳定 `code`、可展示 `message` 和可选 `details`。

环境快照在启动时固定，`revision` 为 `"1"`；工具链快照的 `revision` 随检测递增，设置 revision 随成功写入递增。

JSON 字段为 camelCase；标识符、revision 和契约版本以字符串传输；时间使用 UTC RFC 3339。Rust 当前没有自动生成 TypeScript 类型，`src/features/**/contracts.ts` 由人工保持一一对应。

工具链内部模型不派生 Serde；`ipc/toolchain.rs` 显式映射独立 DTO，领域字段和传输字段不共享类型。

当前业务错误码包括 `TOOLS_DIRECTORY_UNAVAILABLE`、`TOOLS_DIRECTORY_UNSAFE`、`TOOLS_SCAN_FAILED`、`SETTINGS_CONFLICT`、`SETTINGS_CORRUPT`、`SETTINGS_VERSION_UNSUPPORTED`、`PROXY_CONNECTION_FAILED` 和 `PROXY_AUTH_UNSUPPORTED`；安装项诊断包括 `SYSTEM_PROBE_FAILED`、`VERSION_OUTPUT_UNRECOGNIZED` 和 `INSTALLATION_DAMAGED`。前端传输层另有 `IPC_TRANSPORT_FAILED`、`DESKTOP_UNAVAILABLE` 和 `STALE_SNAPSHOT`。

## 工具插件

`ToolPlugin` 当前契约：

| 能力 | 说明 |
| --- | --- |
| `descriptor` | 稳定 `toolId`、显示名、说明和契约版本 |
| `executableNames` | 平台相关可执行文件名 |
| `preferredDirectories` | 登录 PATH 不可用时回退检查的显式环境目录 |
| `knownDirectories` | 插件声明的常见绝对目录 |
| `versionArgs` | 版本探测参数 |
| `parseVersionOutput` | 工具专有输出解析 |
| `sourceHint` | 根据规范路径生成来源提示 |

`lib.rs` 显式创建 `PluginRegistry` 并注册 `UvPlugin`、`PythonPlugin`、`NodePlugin`、`NpmPlugin`、`PnpmPlugin`、`GitPlugin`、`BashPlugin` 和 `ZshPlugin`；插件接收 OS、用户目录和默认 Shell 登录 PATH 等最小探测上下文，而不依赖完整环境模型。通用服务按登录 PATH 顺序选择每个工具的第一个可执行文件，负责路径规范化、只读进程探测、清单读取、当前安装判定和领域摘要，IPC 边界再映射响应 DTO。服务不包含具体工具的版本输出解析规则。

Python 的当前解释器同样由登录 PATH 决定，来源提示区分虚拟环境与 Conda；只有无法取得默认 Shell 登录 PATH 时才回退到显式环境和常见目录。Bash、Zsh 使用账户记录中的默认 Shell 路径，其他 Shell 不产生当前系统安装。版本目录、安装、验证和激活通过后续插件能力与 `operations` 服务接入；`operations` 只使用 `toolchain` 公开的能力接口。

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

代理设置写入 `.devbox/config.json`，当前使用 `schemaVersion: 2` 并兼容读取版本 1；写入使用同目录临时文件、原子替换和 `config.json.bak` 备份，保存前先验证现有文件，损坏内容不会覆盖有效备份。Unix 文件权限为 `0600`。当前没有写入托管 manifest、active 指针或 `.devbox/bin` 的代码，也没有安装库存的备份迁移或跨进程锁。

安装链路接入后，托管安装清单与 active 指针的唯一写入入口归 `toolchain` 库存服务；`operations` 提交验证通过的安装计划，不直接写这些文件。

## 并发与缓存

- `bootId` 在每次 Rust 进程启动时生成。
- 工具链 revision 和最新快照缓存归 `ToolchainService`，revision 在检测开始时分配并在进程内单调递增。
- 并发刷新不会让较低 revision 覆盖已缓存的较高 revision。
- 当前系统安装发现和托管安装扫描通过 `spawn_blocking` 脱离异步执行器，避免阻塞文件操作占住 Tokio 工作线程。
- 系统代理读取通过 `spawn_blocking` 执行；设置服务将网络适配结果转换为设置模块模型，IPC 不依赖基础设施类型。
- 设置更新由单一写入线程串行完成版本检查、重新验证和文件写入；锁只保护短时内存快照，`expectedRevision` 阻止旧窗口覆盖新设置。
- 前端快照 Store 合并同一业务查询；Rust 单飞探测、任务注册表和跨进程锁随操作服务建立。
- 前端断连或被销毁不影响当前只读检测；后台写任务归 `operations` 状态机。

## 日志与安全边界

- 日志插件在 `.devbox` 初始化成功后启用，目标为 `.devbox/logs/devbox.log`，单文件上限 10 MiB。
- 版本命令使用绝对可执行路径和参数数组，不使用 Shell；清空环境后仅保留绝对 PATH 条目以及 Windows 运行所需的最小系统变量，使 npm/pnpm 等 Node.js 启动脚本能够解析解释器，同时避免从相对目录执行同名程序。
- 读取默认 Shell 登录 PATH 时会执行该 Shell 的固定 `-l -c` 打印命令，因此会加载用户的登录启动文件；该命令不接收前端输入，并有 5 秒超时和输出上限。
- 版本探测有 3 秒超时、stdout/stderr 各 64 KiB 上限，并终止直接子进程。
- PATH 中的相对目录会被跳过，避免执行当前工作目录中的同名程序。
- 系统模式使用 reqwest 读取操作系统静态 HTTP/HTTPS 代理，Linux 无统一系统代理时使用代理环境变量；设置快照使用同一代理匹配器返回实际生效路由，按最终选中路由的地址与认证信息区分环境变量与系统设置来源，并只暴露脱敏 URL 和认证需求。直连和手动模式显式关闭自动代理，手动路由只使用保存的主机和端口。目的地强制 HTTPS，重定向最多 5 次。
- 下载任务应在开始时捕获当次网络路由，执行期间保持同一代理；设置变更只影响之后开始的任务。
- 代理测试只访问固定官方目标，不接受前端传入任意 URL；代理不要求账号密码，407 返回 `PROXY_AUTH_UNSUPPORTED`。
- `build.rs` 为六个自定义 command 生成 allow/deny ACL；`capabilities/default.json` 只向 `main` 窗口放行这些命令。

## 构建与配置

| 文件 | 当前配置 |
| --- | --- |
| [Cargo.toml](Cargo.toml) | edition 2024、Tauri 2、Serde、Tokio、UUID、time、reqwest 和日志依赖 |
| [build.rs](build.rs) | Tauri build 和自定义 command AppManifest |
| [tauri.conf.json](tauri.conf.json) | DevBox 标识、主窗口、开发地址、前端产物和打包配置 |
| [capabilities/default.json](capabilities/default.json) | main 窗口的 `core:default` 与六个业务命令权限 |
| `tauri.<platform>.conf.json` | 各平台打包目标；没有实际原生验证前不代表已支持 |

当前 CSP 仍为 `null`。没有远程内容，但正式发布前仍需配置并验证收紧后的 CSP。

## 后续交付

- Release 目录、资产选择、下载、摘要校验、解压、安装、更新和取消进入工具安装链路。
- 操作任务、channel、event、重连查询、启动恢复和跨进程锁进入 `operations` 聚合。
- 持久化格式迁移、备份恢复、卸载、缓存清理和修复命令进入完整工具管理能力。
- 代理认证、镜像和离线目录进入后续网络设置；当前代理支持系统、直连和手动 HTTP/SOCKS5H。
- Python 复用 uv 托管运行时；Node.js 和 Git 作为独立插件接入。
- Windows 和 Linux 在目标平台完成原生构建与运行验证。
- TypeScript IPC 类型和运行时契约校验由契约生成流程维护。
