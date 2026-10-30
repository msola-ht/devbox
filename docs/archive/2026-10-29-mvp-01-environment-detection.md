# MVP 0.1 第一步实施与归档：环境检测

[架构设计](../architecture.md) · [开发指南](../development.md) · [文档索引](../README.md)

本文是历史开发日志，记录当时的实施、审查和验证证据；当前架构决策以[架构设计](../architecture.md)为唯一基线。

状态：已完成并归档

日期：2026-10-29

## 目标

在 macOS arm64 上建立 DevBox 的第一条真实业务链路：通过类型化 Tauri IPC 读取系统环境，创建用户目录 `.devbox`，只读检测系统 uv，并在前端展示系统环境、`.devbox` 路径和 uv 安装状态。

本步骤用于验证前端页面、业务模块、IPC DTO、Rust 应用服务、工具插件和只读进程探测之间的边界，不进入下载和安装。

## 范围

- 建立 `environment`、`toolchain` 和 `ipc` 模块骨架。
- 建立编译期工具插件注册表，首个插件为 `uv`。
- 实现 `get_environment`、`list_tools` 和 `refresh_tools`。
- 首次启动创建用户主目录下的 `.devbox` 根目录。
- 检测 PATH 和常见 macOS 安装位置中的 uv，执行版本命令并解析版本。
- 扫描 `.devbox/tools/uv` 下的托管安装清单；当前没有托管安装时返回空列表。
- 前端建立 `environment`、`toolchain` 业务模块和对应 API 适配。
- 控制台和工具页展示加载、失败、系统安装、未安装和刷新状态。
- 替换示例项目导航为概览、工具和设置入口。

## 非范围

- 下载、校验、解压、安装、更新、取消和卸载。
- 修改 active 指针或生成 `.devbox/bin`。
- Python、Node.js 和 Git 插件。
- 操作任务 channel、event、跨进程锁和完整持久化迁移。
- Windows、Linux 原生验证。
- 新增测试文件或修改现有测试资产。

## 开发步骤

1. 冻结本次 IPC DTO、稳定错误码和 TypeScript 契约。
2. 建立 Rust `state`、`ipc`、`environment`、`toolchain`、`infrastructure/path` 和 `uv` 插件模块。
3. 在应用启动时解析用户主目录、建立 `.devbox` 根目录并保存只读环境快照。
4. 实现 uv 系统探测、托管清单扫描和工具快照缓存。
5. 注册 Tauri command，使用 `bootId + revision` 返回快照。
6. 安装前端 Tauri API 依赖，建立业务 API 适配和查询 hook。
7. 组合概览页和工具页，处理加载、失败、空结果、系统安装和刷新。
8. 更新侧栏、路由、页面标题和模块文档。
9. 执行构建、Rust 检查和真实桌面启动验证。
10. 进行全链路审查，修复发现的问题。
11. 将本计划、实施记录、审查结论和限制归档。

## 契约草案

```text
EnvironmentSnapshot
  bootId, revision, generatedAt, os, arch, homeDir, devboxRoot, devboxReady, devboxError

ToolListSnapshot
  bootId, revision, generatedAt, tools[]

ToolSummary
  toolId, displayName, description, pluginContractVersion,
  systemInstallations[], managedInstallations[], activeInstallationId

InstallationSummary
  installationId, version, path, source, sourceHint, healthy, errorCode

IpcError
  code, message, details
```

`revision` 和标识符使用字符串，`generatedAt` 使用 RFC 3339。命令返回的 `bootId` 必须来自同一个 Rust 进程。

## 验收标准

- 当前 macOS arm64 能显示系统、架构和 `<user-home>/.devbox` 路径。
- 首次启动后 `.devbox` 根目录存在，检测过程不修改系统 uv。
- PATH 为空或没有 uv 时，仍能发现 `/opt/homebrew/bin/uv` 等已知位置。
- 当前机器显示系统安装 uv 0.13.1，来源提示为 Homebrew 或系统路径。
- 系统安装与托管安装分开返回，空托管列表显示为“未安装 DevBox 版本”而不是查询失败。
- 刷新命令更新工具快照，旧请求响应不能覆盖新状态。
- Rust 和 TypeScript 契约字段一致，组件不直接拼接 command 名。
- `bun run build`、`cargo check` 和桌面启动通过。
- 不在 `.devbox` 之外创建业务数据，不修改 Shell 配置。

## 全链路审查清单

- 前端页面是否只组合业务模块公开接口。
- 组件是否通过 hook/API 适配调用 IPC。
- Rust command 是否只做边界转换和调用服务。
- uv 专有检测规则是否留在插件，公共路径与进程能力是否独立。
- 系统安装是否保持只读，托管安装是否来自 manifest。
- DTO 字段、错误码和时间/标识符格式是否一致。
- `bootId + revision` 是否在各层保持一致。
- `.devbox` 创建是否有权限和错误处理。
- 加载、失败、空结果和刷新状态是否可区分。
- 文档、当前模块说明和实际代码是否一致。

## 实施结果

本切片已建立第一条可编译的前后端链路，实现范围与计划一致。

Rust 后端：

- `lib.rs` 是组合根，装配 `.devbox` 路径、环境服务、工具链服务、插件注册表和三个 command。
- `get_environment`、`list_tools`、`refresh_tools` 均已注册，并在 Tauri 应用 manifest 中生成命令 ACL；`capabilities/default.json` 仅对 `main` 窗口放行这三个命令和 `core:default`。
- `EnvironmentService` 保存启动时解析的 OS、架构、用户目录和 `.devbox` 状态。
- `ToolchainService` 扫描 PATH 与插件声明的已知目录，规范化路径去重，以绝对路径执行版本命令并解析结果。
- `UvPlugin` 独占 uv 的执行文件名、平台目录、版本参数、输出解析和来源提示；通用服务不再写死 uv 参数。
- 插件描述包含稳定 `toolId`、显示名、说明和 `contractVersion`；托管清单校验插件 ID 与契约版本。
- 托管安装扫描拒绝符号链接工具根目录、越界清单路径和逃逸安装目录的可执行文件；active 指针校验 `schemaVersion`、`toolId` 和安装存在性。
- 工具快照 revision 在检测开始时分配；Rust 缓存拒绝更旧 revision 覆盖。
- `.devbox` 根目录和全部子目录在 Unix 下统一设为 `0700`，并拒绝使用已存在的符号链接目录。
- 调试与发布日志目标均为 `.devbox/logs/devbox.log`，单个日志文件上限 10 MiB；不再写入操作系统日志目录。

前端：

- 建立 `features/environment` 与 `features/toolchain`，页面只通过公开出口使用 hook 和组件。
- `tauri-ipc.ts` 是唯一业务命令传输适配；普通浏览器运行时返回明确的 `DESKTOP_UNAVAILABLE`，不伪造数据。
- 快照 store 区分 `idle`、`loading`、`ready`、`error`；刷新失败保留旧数据并显示错误，旧 `revision` 不覆盖新快照。
- 初次查询使用 `list_tools`，显式刷新使用 `refresh_tools`，环境查询使用 `get_environment`。
- 概览与工具页展示加载、失败、成功空注册表、系统安装、托管安装、损坏安装和 active 状态。
- 工具说明来自插件 DTO，组件不再写死 uv 文案；导航替换为概览、工具和设置。

## 全链路审查与修复

审查沿“页面 → feature hook → API 适配 → Tauri command → 应用服务 → 插件 / 基础设施 → DTO 返回展示”完成，并核对相邻写入者、缓存和错误路径。

| 发现 | 影响 | 处理 |
| --- | --- | --- |
| 并发检测完成后才分配 revision，缓存可能被较旧结果覆盖 | 权威工具快照可能倒退，重新查询得到旧状态 | 检测开始时分配 revision；Rust 缓存仅接受更高 revision |
| 系统探测参数与日志文案写在通用服务 | uv 规则泄漏到公共层，新增插件需修改公共代码 | `versionArgs` 归插件，日志使用插件 `toolId` |
| 工具卡片写死 uv 说明和空状态 | 第二个插件接入即展示错误文案 | 描述进入 `ToolDescriptor` / DTO，卡片只消费数据 |
| 托管清单路径与可执行文件未做越界校验 | 恶意或损坏清单可能引用安装目录外文件 | 限制纯相对路径，并规范化校验目标仍在安装目录内 |
| active 指针未校验格式版本和工具归属 | 损坏指针可能被当成当前安装 | 校验 `schemaVersion`、`toolId`、安装存在且健康 |
| Unix 只收紧 `.devbox` 根目录 | 子目录仍可能为 `0755` | 所有目录逐个校验并设为 `0700` |
| 业务日志使用系统日志目录 | 日志不在 `.devbox`，与用户目录约定冲突 | 日志目标改为 `.devbox/logs` |
| 自定义 command 未生成 ACL | command 边界依赖默认行为，权限模型不完整 | 增加 AppManifest 和三项窗口级 allow 权限 |
| 普通浏览器暴露底层 `invoke` TypeError | 用户无法判断是运行环境问题 | 明确返回 `DESKTOP_UNAVAILABLE` 和可展示文案 |
| 工具页失败后目录徽标仍显示“检查中” | 失败状态被误判为持续加载 | 增加“不可用”和“目录异常”分支 |
| 刷新失败且已有环境快照时没有提示 | 用户可能把旧数据当成新结果 | 保留旧数据并显示刷新错误 |
| 系统版本命令成功但输出无法解析仍标记健康 | 展示“版本未知”却显示健康 | 标记 `VERSION_OUTPUT_UNRECOGNIZED`，健康状态为否 |
| IPC command 持有缓存、revision 和检测日志编排 | 工具链策略变化需要同时修改边界层和状态层 | 缓存、revision 和日志收回 `ToolchainService`，IPC 只做边界映射 |
| 领域摘要直接派生 Serde 并作为线上 DTO | 领域字段和传输契约共享错误的变化原因 | 领域模型移除 Serde，IPC 定义独立 DTO 并显式映射 |
| 插件注册表内部固定构造 `UvPlugin` | 新增插件或替换探测实现需要修改注册表内部 | 注册表接收插件集合，由 `lib.rs` 组合根显式注册 |
| 插件依赖完整环境模型，目录语义散落在服务中 | 插件接口过宽，路径调整需要跨模块修改 | 插件改用最小探测上下文；`DevboxLayout` 统一目录语义 |
| 概览页和工具页重复跨模块刷新与加载状态 | 任一刷新策略变化需要同步修改两个页面 | 新增应用组合层，统一联合加载和刷新 |
| 查询进行中触发刷新会被静默合并为普通查询 | 显式刷新可能没有执行重检测 | 刷新在当前查询结束后排队，共享同一排队任务 |
| 系统探测失败仍显示“检测到系统版本” | 异常系统安装被误报为有效检测结果 | 状态派生移入模块模型，健康系统版本与异常版本分开 |
| 系统与托管安装可能共享展示 ID | active 可能标记串到系统安装，React key 重复 | active 仅匹配健康托管安装，展示键包含来源和路径 |
| 异步服务直接执行目录扫描和路径规范化 | 阻塞 I/O 可能占用 Tokio 工作线程 | 候选发现与托管扫描迁移到 `spawn_blocking`，异常映射为 `TOOLS_SCAN_FAILED` |

修复后重新执行静态检查与桌面编译，均通过。

## 验证证据

静态与构建：

- `cargo check`：通过，无 warning。
- `bun run build`：通过，TypeScript 与 Vite 产物生成成功。
- `bun run lint`：通过。
- `cargo clippy --all-targets -- -D warnings`：通过。
- 本次修改的 TypeScript/TSX 文件通过 Prettier 检查。
- `bun tauri build --debug --no-bundle`：通过，产出 `src-tauri/target/debug/devbox`。

前端页面：

- 使用独立浏览器打开 `#/` 和 `#/tools`，概览、工具、设置导航和 Breadcrumb 正常。
- 浏览器不具备 Tauri IPC 时，两个页面均显示 `DESKTOP_UNAVAILABLE`，工具目录徽标显示“不可用”，未伪造成功或空列表。
- 页面视口、侧栏、内容区和错误卡片未出现外溢或重复壳层。

本机只读证据：

- 系统：macOS 15.8.1，arm64。
- uv：`uv 0.13.1 (Homebrew 2026-10-14 aarch64-apple-darwin)`，命令路径 `/opt/homebrew/bin/uv`，规范目标位于 `/opt/homebrew/Cellar/uv/0.13.1/bin/uv`。
- Node.js `v24.22.1`、Git `2.56.0`、Python `3.12.14` 可见，但本切片未注册对应插件。
- `.devbox` 与当前 12 个子目录实测均为 `0700`，所有者为当前用户。

未观察项：

- 最终版本的桌面窗口和 Tauri IPC 成功响应未在自动化界面中观察；当前环境未获得 DevBox 原生窗口的 Computer Use 授权，独立浏览器也无法调用本机 IPC。
- 因此“本机 uv 0.13.1”是 shell 只读检测证据，不等同于最终桌面 UI 已显示该结果。
- 已在 `.devbox/logs/devbox.log` 观察到当前 `ToolchainService` 的工具检测成功记录；最终桌面窗口中的页面展示仍未授权观察。

## 残余限制

- 本切片只提供检测，不提供 Release 目录、下载、校验、解压、安装、更新、取消或卸载。
- 插件契约目前只覆盖描述、系统检测和版本解析；`supports`、`list_releases`、`select_asset`、`install_plan`、`validate` 和激活计划尚未实现。
- 没有操作任务注册表、channel、event、跨进程文件锁、启动恢复和状态迁移。
- Windows 与 Linux 只有代码分支准备，没有原生构建或运行验证。
- CSP 仍为 `null`；当前 WebView 没有远程内容，但正式发布前必须收紧。
- active 指针损坏和托管清单损坏会降级为无 active 或“托管版本异常”，尚未提供独立诊断错误码和修复入口。
- 系统探测只终止直接子进程；子进程树清理、整体输出读取时限和更完整的 Windows 运行环境仍需在后续任务实现中验证。
- 未新增或修改测试文件；本次验证由静态检查、构建、源码链路审查和页面观察组成。
- 全量 `src/**/*.{ts,tsx}` Prettier 检查仍会报告 5 个模板或 CLI 生成文件；这些文件不在本切片修改范围内，未为通过检查而改写，本次改动文件已单独验证通过。

## 后续阶段

1. 在可观察的 macOS 桌面环境完成最终 IPC 用户路径复核，记录 `bootId`、revision、uv 版本和来源提示。
2. 补齐 `.devbox` 版本化清单、active 指针写入、原子替换和跨进程锁。
3. 建立操作任务注册表、channel 进度、终态 event、重连查询与启动恢复。
4. 扩展 uv 插件的官方目录、平台资产选择、摘要校验和安装计划。
5. 完成下载、安全解压、版本验证、激活、取消和第二个版本回滚路径。
6. 在 Windows x64 与 Linux glibc x64 做原生验证后，再声明多平台发布支持。
