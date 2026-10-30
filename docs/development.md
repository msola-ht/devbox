# DevBox 开发指南

[文档索引](README.md) · [架构设计](architecture.md) · [前端模块说明](../src/README.md) · [后端模块说明](../src-tauri/README.md)

本文是 DevBox 开发指南，依据[唯一架构基线](architecture.md)描述当前实现视图、后续实施顺序、模块接入规范和验证方式。架构决策以架构文档为准。

## 当前实现视图

DevBox 采用 React 前端、Tauri IPC 和 Rust 桌面后端组成的单应用架构。

```text
main.tsx
  ThemeProvider
    RouterProvider
      AppLayout
        pages
          app 组合层
          feature hooks / components
          feature API 适配
          lib/tauri-ipc
          Tauri IPC
            ipc DTO / error
            application services
            tool plugins
            infrastructure
```

当前业务链路提供环境查询、工具状态查询、显式刷新、代理设置和代理连接测试。

```text
DashboardPage / ToolsPage
  → useToolchainWorkspace
  → useEnvironment / useToolchain
  → getEnvironment / listTools / refreshTools
  → get_environment / list_tools / refresh_tools
  → EnvironmentService / ToolchainService
  → uv、Python、Node.js、npm、pnpm、Git、Bash 与 Zsh 检测插件
  → DevboxLayout / probe_version

SettingsPage
  → useSettings
  → getSettings / updateSettings / testProxy
  → get_settings / update_settings / test_proxy
  → SettingsService
  → config.json
  → hyper-util 系统与环境代理匹配器
```

工具链快照和代理设置由 Rust 持有权威结果，前端保存查询快照和表单草稿。环境与工具链模块通过 `src/app` 协调联合加载和刷新，页面只负责路由与展示组合。

## 模块职责

| 层 | 职责 |
| --- | --- |
| `pages` | 路由入口和页面布局组合 |
| `src/app` | 跨业务模块的前端状态协调 |
| `src/features` | 单个业务能力的 API、契约、状态、模型和组件 |
| `src/components` | 业务无关的壳层、页面容器和基础 UI |
| `src/hooks` | 业务无关的通用前端能力 |
| `src/lib` | 传输和通用工具 |
| `ipc` | command、响应 DTO、请求 DTO 和稳定错误码 |
| `modules` | 业务模型、用例服务和模块公开接口 |
| `plugins` | 单个工具的探测、目录、资产、安装和激活规则 |
| `infrastructure` | 进程、路径、下载、校验、解压和持久化能力 |
| `lib.rs` | 服务、插件、基础设施和 command 的唯一组合根 |

依赖方向固定为：

```text
pages → app / feature
feature → shared
ipc → application service
application service → domain
plugin → tool contract
application service → infrastructure
operations → toolchain public interface
lib.rs → all composition
```

业务模块通过 `index.ts` 或模块公开出口对外提供能力。页面、组件和状态容器通过 feature API 适配层访问 IPC。`operations` 只通过 `toolchain` 公开出口取得库存与安装能力，`toolchain` 不依赖 `operations`。

## 当前工具契约

系统检测插件实现统一的 `ToolPlugin` 契约：

| 能力 | 内容 |
| --- | --- |
| `descriptor` | 稳定工具 ID、显示名称、说明和契约版本 |
| `executable_names` | 平台相关可执行文件名 |
| `preferred_directories` | 登录 PATH 不可用时回退检查的显式环境目录，例如当前 Python 虚拟环境 |
| `known_directories` | 工具常见安装目录 |
| `version_args` | 版本探测参数 |
| `parse_version_output` | 工具专有版本输出解析 |
| `source_hint` | 来源提示 |

`lib.rs` 组合 `UvPlugin`、`PythonPlugin`、`NodePlugin`、`NpmPlugin`、`PnpmPlugin`、`GitPlugin`、`BashPlugin` 和 `ZshPlugin`。工具名称、平台目录、版本解析和来源提示归插件；默认 Shell 登录 PATH 解析、路径规范化、进程探测、当前安装判定、托管清单扫描、快照缓存和 revision 归工具链服务。只有无法取得登录 PATH 时才回退到 `VIRTUAL_ENV`、`CONDA_PREFIX` 和常见目录。

## 后续实施顺序

### 工具能力契约

工具能力按生命周期拆分为描述、系统检测、Release 查询、资产选择、安装计划、验证和激活。注册表保存工具描述及可用能力，`toolchain` 使用检测能力并通过公开接口向 `operations` 提供安装相关能力。

```text
ToolRegistration
  descriptor
  detector
  release_source?
  install_planner?
  validator?
  activation_planner?
```

插件返回声明式计划和类型化结果。`operations` 负责下载、解压、临时文件、任务状态和失败恢复；验证通过后通过 `toolchain` 公开库存接口写入安装清单、active 指针和 `.devbox/bin`。

平台探测上下文包含 OS、架构、libc、用户目录和必要的运行环境信息。工具插件根据显式平台能力选择支持范围、Release、资产和安装计划。

### 前后端契约

Rust IPC DTO 作为契约源，TypeScript 请求、响应、枚举和错误详情由生成流程产出。生成物由命令维护，前端业务模块消费生成类型。

每个 command 保持单一职责：

```text
query command    → 读取权威状态
action command   → 提交意图并返回资源或任务标识
stream/channel   → 传输有顺序的进度或流数据
event            → 传输低频状态变化通知
```

查询快照包含 `bootId`、业务作用域和单调 revision。前端按业务作用域、资源 ID、筛选和分页组织缓存键。

### 操作与任务

安装、更新、切换、取消、恢复、卸载和任务查询归 `operations` 聚合。托管安装清单与 active 状态归 `toolchain`；`operations` 通过公开库存接口提交最终写入。

```text
modules/operations/
  mod.rs
  model.rs
  service.rs
  error.rs

ipc/operations.rs

src/features/operations/
  index.ts
  contracts.ts
  api.ts
  hooks/
  components/
```

首个操作链路使用以下契约：

```text
start_install    → OperationAccepted
get_operation    → OperationSnapshot
cancel_operation → OperationSnapshot
```

`start_install` 返回已登记的操作 ID 和初始状态。进度通过带 `operationId` 和单调序号的 channel 传输；重新进入页面后通过 `get_operation` 恢复权威状态。

任务状态与执行阶段相互独立。状态采用以下单调状态机：

```text
queued → running
queued | running → cancel_requested
running → succeeded | failed
cancel_requested → cancelled
queued | running | cancel_requested → interrupted
```

`succeeded`、`failed`、`cancelled` 和 `interrupted` 是终态，进入终态后不回退。`interrupted` 由启动恢复在识别进程异常退出后写入。

执行阶段只在 `running` 状态内推进：

```text
resolving → downloading → verifying → extracting
          → validating → activating → finalizing
```

操作请求携带 `clientRequestId`，重复提交映射到同一个任务。任务注册表、跨进程锁和持久化记录由 Rust 管理。

### 安装执行链

首个完整安装链路以 uv 为对象：

```text
operations service
  → 获取工具安装能力
  → 生成 InstallPlan
  → 下载到 downloads
  → 校验大小与摘要
  → 安全解压到 staging
  → 插件验证产物
  → 向 toolchain 提交验证产物与安装计划
  → toolchain 原子写入安装库存
  → toolchain 更新 active 指针与 .devbox/bin
  → 发布完成状态
```

通用执行器负责下载、验证、重试边界、取消协作、临时资源清理和失败补偿。插件负责工具专有的 URL、摘要、归档结构、可执行文件路径、验证参数和激活规则。`toolchain` 是安装清单、active 指针和命令入口的唯一写入者。

### 工具扩展

uv 安装链路完整后，按能力复用顺序扩展：

1. Python 复用检测契约，并提供 uv 托管 Python 的安装、验证和激活计划。
2. Node.js 验证归档安装、校验文件和双版本切换。
3. Git 验证平台化检测、系统权限边界和安装策略。
4. 操作历史、缓存和多工具查询形成独立业务模块。
5. 代理设置已建立系统、直连、手动 HTTP/SOCKS5H、持久化和连接测试；继续增加代理认证、镜像、离线缓存和 Shell 集成。
6. 操作记录和缓存规模达到结构化查询需求后，引入持久化迁移与 SQLite。
7. 第三方插件在独立进程或受限 WebAssembly 沙箱中运行。

## 接入新工具检测

工具检测按插件接入：

1. 在 `src-tauri/src/modules/toolchain/plugins/<tool>/` 建立插件模块。
2. 定义稳定的 `toolId`、显示名称、说明和契约版本。
3. 实现平台目录、可执行文件名、版本参数、版本解析和来源提示。
4. 在 `lib.rs` 注册插件。
5. 通过 `list_tools` 和 `refresh_tools` 返回统一工具快照。

通用工具卡片、快照 store、IPC DTO 和页面保持共享。工具检测的变更范围集中在插件目录和组合根。

## 接入新查询资源

新的业务查询建立完整垂直切片：

```text
modules/<feature>/      业务模型和应用服务
ipc/<feature>.rs        command 和 DTO
features/<feature>/     前端 API、契约、状态和组件
pages/<feature>.tsx     路由页面
```

查询服务拥有资源读取、过滤、分页、权限和响应映射。feature hook 拥有查询缓存和交互状态。页面拥有路由参数和布局组合。

跨 feature 协调进入 `src/app`。模块之间通过公开出口和业务标识协作，不读取对方内部 store、hook 或组件实现。

查询按业务资源建模块，不建立集中所有资源的通用查询服务；任务查询留在 `operations`，工具与安装库存查询留在 `toolchain`。

## 接入安装能力

安装能力沿以下边界实现：

1. 插件提供 `ReleaseSource`、`InstallPlanner`、`Validator` 和 `ActivationPlanner`。
2. `operations` 服务校验工具 ID、版本、平台、并发状态和客户端请求 ID。
3. `operations` 创建任务并调用通用执行器。
4. 基础设施提供下载、摘要校验、安全解压、原子写入适配和清理能力。
5. 插件验证工具专有产物。
6. `operations` 将验证产物和安装计划提交给 `toolchain` 公开库存接口。
7. `toolchain` 写入安装清单、active 指针和命令入口。
8. 前端 `operations` feature 提交意图、接收进度并查询最终状态。

工具页面提交业务意图，不接触下载地址、临时路径、命令名或 Tauri 原始消息。

## 状态与并发

- `toolchain` 持有工具、安装清单、激活状态及其 revision。
- `operations` 持有操作记录、任务状态和进度序号。
- 前端保存查询缓存、筛选条件和编辑草稿。
- 同一任务只有一个状态所有者，状态按 `operationId` 隔离。
- 旧作用域响应通过作用域和 revision 丢弃。
- 刷新期间保留已有数据，并单独表达刷新错误。
- 写操作通过 `clientRequestId` 去重，通过工具锁串行化同一工具的写事务。
- 阻塞文件和进程准备进入 `spawn_blocking`，CPU 密集计算使用受控工作线程。
- 锁覆盖短时一致性更新，不跨越外部 I/O。

## 开发环境

```bash
bun install
bun tauri dev
```

仅调试前端：

```bash
bun run dev
```

## 多平台构建

Tauri 按当前操作系统加载平台配置。

| 平台 | 命令 | 产物 |
| --- | --- | --- |
| macOS | `bun tauri build` | `.app`、`.dmg` |
| Linux | `bun tauri build` | AppImage、`.deb`、`.rpm` |
| Windows | `bun tauri build` | NSIS `.exe`、MSI |

macOS 通用二进制：

```bash
rustup target add aarch64-apple-darwin x86_64-apple-darwin
bun run tauri:build:macos:universal
```

Ubuntu 构建依赖：

```bash
sudo apt update
sudo apt install libwebkit2gtk-4.1-dev build-essential curl wget file libxdo-dev libssl-dev libayatana-appindicator3-dev librsvg2-dev
```

发布构建在目标平台补充 macOS 公证、Windows 代码签名和 Linux 包签名。

## UI 组件

shadcn/ui 组件通过官方 CLI 生成：

```bash
bunx --bun shadcn@latest add button
```

覆盖已有组件或 block 前先预览：

```bash
bunx --bun shadcn@latest add <component-or-block> --dry-run
bunx --bun shadcn@latest add <component-or-block> --overwrite
```

`src/components/ui/**` 保持 CLI 生成结构。业务页面通过公开 API、variants 和 props 组合组件。

不主动使用 Computer Use。只有用户明确要求桌面或浏览器 UI 自动化时才调用；未获授权时完成可执行的构建和静态检查，并明确说明未能观测的 UI 行为。

## 验证

前端构建和类型检查：

```bash
bun run build
bun run lint
```

Rust 静态检查：

```bash
cargo check --manifest-path src-tauri/Cargo.toml
cargo clippy --manifest-path src-tauri/Cargo.toml --all-targets -- -D warnings
```

桌面构建：

```bash
bun tauri build --debug --no-bundle
```

UI 和路由修改通过浏览器检查布局、主题、空状态、失败状态和视口滚动。涉及 IPC 或 Rust 业务行为时，通过 `bun tauri dev` 观察真实桌面操作从入口到权威结果。浏览器页面验证前端表现，桌面运行验证 IPC、目录和进程链路。

通信机制参考 Tauri 官方文档：[前端调用 Rust](https://v2.tauri.app/develop/calling-rust/)、[Rust 调用前端](https://v2.tauri.app/develop/calling-frontend/)和[能力权限](https://v2.tauri.app/security/capabilities/)。
