# Devbox 架构

Devbox 使用 MyGo 承载 React 界面和 Go 本机服务。前端负责展示与交互，Go 负责系统访问、业务编排和业务数据持久化，双方通过自动生成的类型化接口通信。

开发与验证遵循 [项目规则](AGENTS.md)，产品目标和待实现功能见 [产品设计](DESIGN.md)及[全局代理模块设计](PROXY_DESIGN.md)。目录入口见 [前端索引](src/index.md) 和 [后端索引](internal/index.md)。各目录的 `index.md` 记录职责及直属文件、子目录入口，通信与行为契约集中在本文。

## 入口与职责

| 位置 | 职责 |
| --- | --- |
| [main.go](main.go) | 分派系统代理查询子命令，组装 Go 服务和退出清理，创建窗口并启动 MyGo 应用 |
| [project_dialog.go](project_dialog.go) | 将项目目录操作接入原生目录选择、文件管理器和 MyGo 绑定 |
| [mygo.config.ts](mygo.config.ts) | 配置前端开发地址、构建命令、产物和接口生成路径 |
| [src/main.tsx](src/main.tsx) | 恢复上次路由，挂载 React、StrictMode、HashRouter 和浮动提示 |
| [src/app.tsx](src/app.tsx) | 将合法路由映射为页面，组合应用外壳、导航和页面 |
| [src/app-routes.ts](src/app-routes.ts) | 定义路由及标题，校验路径，保存与恢复上次页面 |
| [src/features](src/features/index.md) | 按业务组织页面、组件、Hook 和业务内工具 |
| [src/features/runtime-env](src/features/runtime-env/index.md) | 概览页的运行环境展示 |
| [src/features/devtools](src/features/devtools/index.md) | 工具目录、检测卡片、安装详情与本机安装操作 |
| [src/features/settings](src/features/settings/index.md) | 代理设置、配置恢复和连接检查 |
| [src/components/ui](src/components/ui/index.md) | shadcn 基础组件，由业务组件组合使用 |
| [src/lib](src/lib/index.md) | 已被复用的通用前端工具 |
| [src/mygo.ts](src/mygo.ts) | MyGo 生成的 TypeScript 类型与服务调用入口 |
| [internal/systeminfo](internal/systeminfo/index.md) | 当前的系统信息业务，包括采集、快照存储与服务编排 |
| [internal/runtimeenv](internal/runtimeenv/index.md) | 用户主目录和默认登录 Shell 查询 |
| [internal/devtools](internal/devtools/index.md) | 本机开发工具检测与 macOS、Windows、Ubuntu 的安装 |
| [internal/projectdir](internal/projectdir/index.md) | 默认工作目录、项目路径校验与受限的工具目录访问 |
| [internal/proxy](internal/proxy/index.md) | 代理配置、路由解析和连接检查 |
| [internal/configfile](internal/configfile/index.md) | 配置文件定位和 JSON 写入 |
| [internal/localcmd](internal/localcmd/index.md) | 本机命令执行、输出与子进程取消 |
| [internal/readcache](internal/readcache/index.md) | 同项读取协调与会话内结果复用 |

应用外壳通过属性接收侧栏、面包屑和页面内容。页面组合业务组件；导航组织入口与路由信息。新增业务沿用 `src/features/<业务>` 和 `internal/<业务>` 的组织方式。

## 模块边界与依赖方向

```mermaid
flowchart TD
    Entry[应用入口与路由] --> View[页面与业务组件]
    View --> UI[shadcn 基础组件]
    View --> Hook[业务 Hook]
    Hook --> Client[MyGo 生成接口]
    Client --> Service[Go Service]
    Service --> System[系统信息]
    Service --> Environment[运行环境]
    Service --> DevTools[开发工具检测与安装]
    Service --> ProjectDialog[项目目录绑定]
    DevTools --> ProjectDirectory[项目目录]
    ProjectDialog --> ProjectDirectory
    Service --> Proxy[代理设置]
    System --> Collector[系统采集]
    System --> Snapshot[快照存储]
    System --> ReadCache[会话读取缓存]
    Environment --> ReadCache
    DevTools --> ReadCache
    Proxy --> ReadCache
    Proxy --> Network[路由解析与连接检查]
    Proxy --> Settings[代理配置存储]
    Snapshot --> ConfigFile[配置文件定位与 JSON 写入]
    Settings --> ConfigFile
    Collector --> Command[本机命令执行]
    Environment --> Command
    DevTools --> Command
    Network --> Command
```

- 页面负责组合刷新入口，卡片负责展示。[use-refreshable.ts](src/hooks/use-refreshable.ts) 协调系统信息、运行环境和开发工具的首次读取、手动刷新与取消。设置表单管理独立的读取、写入和连接检查。
- Go Service 组合具体采集与存储能力，负责业务顺序、超时和错误处理。被绑定服务的导出方法属于前端可调用契约。
- 采集模块负责系统接口与平台命令，存储模块负责文件格式与读写，二者通过数据模型协作。
- [localcmd](internal/localcmd/index.md) 执行本机命令并管理子进程，业务模块指定命令、环境和上下文，按需要捕获或实时接收输出。Windows 的显卡和磁盘采集共用 PowerShell JSON 查询入口。
- Go 业务服务持有各自的缓存。[readcache](internal/readcache/index.md) 复用最近成功的读取结果，协调同一项的并发读取。业务服务决定有效期与主动刷新的语义。
- 业务界面通过 shadcn 组件和内置变体组合。业务内的格式化和请求逻辑位于所属业务目录。
- 当前采用具体函数组合。公共模块与接口的提取以实际复用、替换实现或依赖解耦需求为依据。

## 通信与执行方式

### 当前通信入口

Go 在 `main.go` 中注册系统信息服务 `SystemService`、运行环境服务 `EnvironmentService`、开发工具服务 `DevToolsService`、代理设置服务 `ProxyService` 和目录操作 `ProjectDialog`。前端通过 `src/mygo.ts` 中生成的接口调用它们。

`src/mygo.ts` 是生成产物，类型与方法签名以 Go 定义为源。修改 Go 对外契约后运行 `bun run generate`，同步调整调用方并检查类型。前端共用生成类型与调用入口。

MyGo 接口生成会执行应用入口直到 `App.Run`。入口中保持服务注册和应用组装；需要读写业务数据的初始化工作放在适当的应用生命周期或业务调用中。

### 同步与异步

前端调用返回 Promise，通过 `await` 等待结果，界面仍可响应。每次 MyGo 绑定调用在独立 goroutine 中执行；Go 服务内部按业务顺序执行，只有独立工作才并行。

概览和开发工具的缓存读取与重新采集共用普通调用。概览的统一刷新发起独立请求，各卡片随各自的结果更新。Go 在系统信息采集中并行查询 GPU 和磁盘。

普通调用是当前默认方式；Channel、Event 和后台任务在出现对应需求时按下表引入。

| 实际需求 | 通信与生命周期约定 |
| --- | --- |
| 查询数据、保存配置、执行一次操作 | 普通调用返回结果或错误；明确成功是否包含持久化完成 |
| 单次调用持续产生日志、输出或进度 | 使用 MyGo Channel，随该次调用结束或取消 |
| 数据变化需要通知其他监听者或窗口 | 使用类型化 Event 通知变化，由所属业务协调查询与状态更新 |
| 页面离开后仍需运行的安装、构建等任务 | 由 Go 管理任务生命周期，启动返回任务标识，并提供状态查询和显式取消 |

通信方式与任务形式按输出形式、取消方式和任务与页面的生命周期关系选择。

### 取消、错误与并发

- 页面内的读取和连接检查使用 `AbortController`。刷新会取消同一项的旧请求，离开页面会取消关联读取。代理保存和恢复继续执行，完成后通过全局通知反馈。
- Go 耗时调用接收 `context.Context`，并传给支持取消的系统接口、外部命令和网络操作。每项采集由业务服务设置期限，等待同项缓存读取时也可取消。
- Go 返回的错误使前端 Promise 拒绝。Hook 将其转换为界面状态，组件展示提示并提供重试；取消的请求保持安静。
- 多个 MyGo 调用可以同时执行。同项快照读取、重新采集和保存串行执行，不同数据项可并行。文件使用临时文件写入后替换。
- `ProxyService` 持有代理配置，用互斥锁协调读取、保存与恢复。连接检查使用每次调用传入的表单草稿。

## 状态与持久化归属

| 数据 | 归属与保存方式 |
| --- | --- |
| 系统信息 | 本机采集为数据来源，`SystemService` 保留本次运行中最近一次采集结果，文件快照供下次启动读取 |
| 当前展示的快照、刷新中和错误 | 查询 Hook 的 React 内存状态 |
| 运行环境 | `EnvironmentService` 的会话内结果，由概览统一刷新 |
| 开发工具检测结果 | `DevToolsService` 的会话内结果，按工具单独刷新 |
| 工具安装任务与最近输出 | `DevToolsService` 按工具保留本次运行中最近的任务与输出，页面负责查询与展示 |
| 代理设置 | `ProxyService` 和 `os.UserConfigDir()/devbox/proxy.json` |
| 系统代理展示结果 | `ProxyService` 的 30 秒内存缓存 |
| 代理表单草稿和连接检查状态 | 设置页面的 React 内存状态 |
| 主题、侧栏展开状态、上次页面、已确认的工作目录 | 前端 `localStorage`，分别由主题、外壳、路由和概览模块管理 |

业务数据的存储格式、兼容性和采集策略由 Go 管理。前端本地存储用于界面偏好。主题在 [index.html](index.html) 中先恢复，再由主题模块切换；调整主题存储约定时需要同时核对这两个入口。

## 系统信息模块

### 后端组成

| 文件 | 职责 |
| --- | --- |
| [model.go](internal/systeminfo/model.go) | 定义 `SystemInfo`、`GPUInfo` 和带采集时间的 `SystemSnapshot` |
| [service.go](internal/systeminfo/service.go) | 提供 `Get`，协调内存、文件快照和重新采集 |
| [collector.go](internal/systeminfo/collector.go) | 采集本机数据，协调独立的 GPU、磁盘查询 |
| [snapshot_store.go](internal/systeminfo/snapshot_store.go) | 解析文件位置、读取和校验快照、写入并替换快照文件 |
| `graphics_<平台>.go` | 通过对应平台的系统命令读取 GPU 信息 |
| `storage_<平台>.go` | 读取物理磁盘总容量，按整块磁盘计数；此处的 storage 指磁盘采集 |

一次系统采集的预算为五秒，其中并行的 GPU 和磁盘查询各有三秒预算。可选项失败或超时仍保留其他字段。数值缺失为 `null`，GPU 列表缺失归一化为 `[]`。整体采集取消或超时时返回错误。

### 接口契约

| 前端调用 | 成功返回 | 行为与错误 |
| --- | --- | --- |
| `SystemService.get(false)` | `SystemSnapshot` | 依次使用内存快照、兼容的文件快照和本机采集。文件读取失败时记录日志并采集 |
| `SystemService.get(true)` | `SystemSnapshot` | 重新采集并更新快照。采集失败时保留已有快照，文件保存失败时记录日志并返回本次结果 |

`SystemSnapshot` 包含 `info` 和 `collectedAt`。`collectedAt` 是采集完成的 UTC 时间，经生成接口表示为字符串。

### 页面加载与刷新

[概览页](src/features/workspace/workspace-page.tsx) 分别读取系统信息和运行环境。一个“刷新概览”按钮触发两项独立查询，各卡片按各自的完成或失败状态更新。首次读取期间显示占位，刷新期间保留已有数据。离开页面时取消两项读取。

### 快照文件

文件位置为 `os.UserConfigDir()/devbox/system-info.json`，macOS 对应 `~/Library/Application Support/devbox/system-info.json`。

当前文件格式版本为 `2`，包含 `version`、`info` 和 `collectedAt`。读取时校验版本、采集时间非零、操作系统和架构，额外的 JSON 字段会被忽略。旧格式或不兼容快照会在下次成功采集并写入后被替换。快照和代理设置通过 `configfile` 定位文件并写入同目录临时文件，完成编码和关闭后再替换目标文件。

## 运行环境

概览页在系统信息右侧显示[运行环境卡片](src/features/runtime-env/environment-card.tsx)，内容较窄时改为纵向排列。`EnvironmentService.get(false)` 使用本次运行中的结果，首次读取时查询用户主目录和默认登录 Shell。刷新概览时调用 `get(true)` 重新读取。

`homeDir` 由 Go 的 `os.UserHomeDir` 读取，桌面端取自进程的 `HOME` 或 `USERPROFILE`。在 macOS 和 Linux 上，`loginShell` 通过 UID 查询系统账号，分别调用 `/usr/bin/dscacheutil` 和 `/usr/bin/getent`。查询期限为两秒。主目录缺失、默认登录 Shell 查询失败或平台不适用时，对应字段为空，页面显示“暂不可用”。

## 开发工具检测

前端的[工具目录](src/features/devtools/tool-catalog.ts) 定义开发工具页的分组、工具标识、名称、检测方法和展示范围。分组按主要用途和语言生态划分，平台限制按工具条目声明。前端通过 MyGo 运行时读取平台，隐藏无适用工具的分组。Go 检测服务负责平台校验和命令定位。

| 开发工具分组 | 当前工具 |
| --- | --- |
| 系统包管理器 | macOS 上的 Homebrew，Windows 上的 WinGet 和 Chocolatey |
| 基础工具 | Git、GitHub CLI、jq、ripgrep |
| Python | Python、uv |
| JavaScript | Node.js、npm、pnpm |
| Shell 与脚本 | macOS 和 Linux 上的 Zsh，Windows 上的 PowerShell |

开发工具按组展示卡片，每组每排最多四张，卡片平分所在行的可用宽度。系统包管理器和 Shell 与脚本同时适用时在页面顶部并排，内容区较窄时上下排列。[工具卡片](src/features/devtools/tool-card.tsx) 通过 [DevToolsService](src/mygo.ts) 中各自的检测方法读取结果。进入页面时优先使用最近一次检测结果，首次读取时执行检测。卡片的刷新图标只重新查询对应工具。开始同一工具的新请求或离开页面时取消旧请求。

Go 在 [devtools](internal/devtools/index.md) 内共用 `ToolDetection`、命令查找和命令执行。[发现与编排](internal/devtools/discovery.go) 按 Devbox 进程的 `PATH` 顺序收集绝对目录中的入口，再加入工具及平台特有的路径。Windows 补查注册表中的当前用户和系统 PATH、WinGet 链接目录、应用执行别名和 Python 官方注册路径。Linux 补查 `/usr/bin`，uv 和 pnpm 也检查用户的 `~/.local/bin`。每项检测最多并行探测四个入口，按发现顺序返回 `ToolDetection[]`。单个入口的失败随结果返回。相同入口只探测一次，指向同一实际路径且版本相同的成功结果只显示首次命中的一条。Python 按探测返回的解释器路径区分环境。

卡片以首个可用结果作为摘要，全部不可用时显示首条结果。标题行显示工具名称、可用版本和操作，名称提示包含工具说明和完整版本。可用时内容行并排显示已识别的安装来源和路径，未找到时显示状态，有已知入口时可在状态提示中查看路径。有多条检测结果时，点击刷新图标旁的数字可打开安装详情。详情按检测顺序列出各条结果，成功条目显示完整版本和已识别的来源，失败条目显示原因。有路径的条目在空间不足时单行截断，悬浮或键盘聚焦可查看完整路径。卡片的路径行沿用同样的展示方式。

Go 根据已知位置及解析符号链接后的实际路径识别来源。[通用来源判定](internal/devtools/source.go) 支持 macOS 系统命令、Apple 开发工具和 Homebrew 默认 Cellar。Windows PowerShell 的系统入口标注为系统自带。Python 按基础解释器归属标注来源，虚拟环境的路径与独立结果保留。Node.js 无法归属时显示「未知」，其他工具在来源字段为空时省略来源。

每项检测有六秒预算，单条命令最多运行五秒。查找入口阶段超时会返回整项错误。探测阶段超时时保留已取得的结果，等待中的入口标为尚未探测。调用方取消时停止探测。Unix 终止命令所在的进程组，Windows 终止命令所在的 Job Object。Windows 后台命令使用无控制台窗口模式。成功结果中的路径来自实际执行的入口或解释器。

- [Git](internal/devtools/git.go) 使用 `PATH` 中的绝对入口。在 macOS 上，入口解析为 `/usr/bin/git` 时，通过 `xcode-select --print-path` 查找当前开发工具目录，直接执行其中的 `usr/bin/git`，避免触发安装提示。
- [jq、ripgrep 和 GitHub CLI](internal/devtools/utilities.go) 分别查找 `jq`、`rg` 和 `gh` 并执行 `--version`。macOS 上还检查 `/opt/homebrew/bin` 和 `/usr/local/bin`。GitHub CLI 检测关闭版本更新检查。jq 和 ripgrep 还检查本次运行中安装任务读取到的 Homebrew 前缀。
- [Zsh](internal/devtools/zsh.go) 在 macOS 和 Linux 上执行 `zsh --version`。应用 `PATH` 之后检查 macOS 的常用包管理器路径以及 `/bin/zsh` 和 `/usr/bin/zsh`。
- [PowerShell](internal/devtools/powershell.go) 在 Windows 上查找 `pwsh`，并检查 PowerShell 7 和系统自带 Windows PowerShell 的默认安装路径。探测在不加载用户配置的非交互会话中读取 `$PSVersionTable.PSVersion`。
- [uv](internal/devtools/uv.go) 执行 `uv --version`，从 `PATH` 查找绝对入口。macOS 上，uv 与 ripgrep 还检查各自的常用安装路径和本次运行中安装任务读取到的 Homebrew 前缀。
- [Python](internal/devtools/python.go) 在 macOS 和 Linux 上查找 `python3`，在 Windows 上查找 `python.exe`。macOS 的 `/usr/bin/python3` 与 Git 共用 Apple 开发工具目录查询。Windows 的应用执行别名通过 `pymanager.exe list --one --format=json` 查询已安装解释器，探测为子进程关闭自动安装选项。探测在隔离模式中读取版本、当前解释器路径及可用的基础解释器路径，并识别 uv 的默认数据目录或 `UV_PYTHON_INSTALL_DIR` 中的解释器。
- [Homebrew](internal/devtools/homebrew.go) 在 macOS 上检测。`PATH` 入口之后，按应用进程架构检查 `/opt/homebrew/bin/brew` 和 `/usr/local/bin/brew`。
- [WinGet 和 Chocolatey](internal/devtools/windows_package_managers.go) 在 Windows 上用 `--version` 检测。WinGet 补查 App Installer 的执行别名。Chocolatey 补查安装目录和 `ProgramData` 下的默认入口。
- [Node.js](internal/devtools/node.go) 与 npm 共用已知版本目录的查找和来源识别。先查应用 `PATH`，再查 macOS 常用安装路径以及 Homebrew 默认及首个成功查询的 Cellar 中的 `node` 与 `node@*`、fnm 的 `node-versions`、nvm 的 `versions/node` 和 nvm-windows 安装目录。同一实际路径和版本经多个入口命中时，保留最先发现的入口。
- [npm 和 pnpm](internal/devtools/javascript.go) 执行 `--version`，优先使用入口同目录下的 Node.js。npm 与 Node.js 共用版本目录查找，pnpm 补查 macOS 常用安装路径和绝对路径 `PNPM_HOME`。Corepack 探测关闭联网、下载提示和项目版本选择，使用已配置或内置的默认版本。成功结果的入口解析到 Corepack 脚本时标注来源。Corepack 因离线而无法获取默认版本时，返回 `failed`、入口路径和 `corepack` 来源，提示先安装所需版本再刷新。pnpm 自身的按项目切换版本也在探测时关闭。Windows 上通过系统 `cmd.exe` 执行 `npm.cmd` 和 `pnpm.cmd` 等批处理入口，其子进程纳入同一个 Job Object。

| `status` | 含义 |
| --- | --- |
| `available` | 命令成功并返回可识别的版本，包含 `path` 和 `version` |
| `not_found` | 在查找范围内未找到入口，或从已知的 Python 启动别名未找到已安装解释器，包含 `message` |
| `failed` | 入口无法使用、命令执行失败或超时、版本格式未识别，或探测因整项期限尚未开始，包含 `message` 及已知的 `path` |
| `unsupported` | 当前平台不适用该项检测，包含 `message` |

未找到入口或平台不适用时，数组包含一条对应状态。检测结果保存在 Go 进程内。刷新期间显示上次结果，收到新结果后更新对应卡片。请求失败时保留上次结果并提示重试。一次请求失败只影响对应卡片。

## 本机工具安装

[Go 安装目录](internal/devtools/install_catalog.go) 定义支持的工具、固定包标识、平台和渠道。开发工具页用一次短调用读取适用渠道，通过[共用安装操作](src/features/devtools/tool-install-action.tsx)启动任务。已有可用版本时展示检测结果。Windows PowerShell 5.1 与 PowerShell 7 可并存，安装操作单独检查是否已有可用的 `pwsh`。Homebrew 自身通过卡片打开官方安装说明。Go [工具安装任务](internal/devtools/tool_install.go) 按工具或项目作用域登记最近任务，立即返回初始状态。各入口将已选定的安装或维护流程交给任务管理器后台执行。

| 前端调用 | 行为 |
| --- | --- |
| `DevToolsService.getToolInstallOptions()` | 返回各工具在当前平台上的安装渠道 |
| `DevToolsService.startToolInstall(tool)` | 启动固定渠道的安装。同一工具已有活动任务时返回该任务 |
| `DevToolsService.startWinGetUpdate()` | 检查并更新当前用户的 WinGet，同一工具已有任务时返回该任务 |
| `DevToolsService.getToolInstall(tool)` | 读取最近任务的标识、状态、是否可取消、下载进度、提示和最近 16 KiB 的合并输出。未启动过时返回 `null` |
| `DevToolsService.cancelToolInstall(tool, id)` | 请求取消标识匹配且尚可取消的任务，返回当前任务快照 |

[Homebrew 安装](internal/devtools/homebrew_install.go) 选用首个可用的 Homebrew，通过绝对路径运行固定的 formula 安装命令。同一个解析后的前缀只运行一个 Devbox 安装任务。Go 将实际入口和 Cellar 位置存入检测模块的会话内路径表。Node.js 和 npm 由 `node@24` 提供，从 formula 的 `opt` 目录复检。子进程采用非交互模式，关闭自动更新、目标包自动升级、清理和已安装依赖方的检查。

[WinGet、APT 和 npm](internal/devtools/package_install.go) 分别使用平台包管理器。WinGet 固定社区源和包标识，跳过升级，由系统呈现安装器和授权。源或软件包需要接受条款时，卡片提示用户到终端查看并处理对应条款。Node.js 和 npm 指定 Wix 安装器。APT 首批支持 Ubuntu，使用固定包名和 `--no-remove`，普通用户通过系统认证代理运行。Ubuntu 的 pnpm 使用首个搭配 Node.js 18 或更新版本的 npm，安装 pnpm 12 到 `~/.local` 并验证 `~/.local/bin/pnpm`。同一渠道一次执行一个本机任务。[Ubuntu uv 安装](internal/devtools/user_uv_install.go) 复用官方发行包下载、SHA-256 核验和解包，在用户工具目录发布 `uv` 和 `uvx`，保留既有入口。应用代理用于 Homebrew、npm 和发行包下载，WinGet 与 APT 使用各自的系统连接设置。

[WinGet 维护](internal/devtools/winget_maintenance.go) 使用系统自带 Windows PowerShell。缺少微软的 `Microsoft.WinGet.Client` 模块时，从官方 PowerShell Gallery 安装到当前用户。安装和修复调用 `Repair-WinGetPackageManager`，手动更新先查询微软最新稳定发布，再比较当前版本，只应用严格更新的版本。目标为当前用户，命令结束后复检 WinGet。缺少可用 WinGet 时，依赖它的工具提示先使用 WinGet 卡片。

[Chocolatey 安装](internal/devtools/chocolatey_install.go) 使用系统 UAC 启动同一份 Windows PowerShell，运行 Chocolatey 官方引导程序并复检版本。所有通过 Windows 包管理器安装工具的任务、WinGet 维护和 Chocolatey 安装共用一个任务槽位。WinGet 版本查询通过应用代理，PowerShell Gallery 及 Windows 原生安装器使用各自的连接设置。

页面在上一次活动状态读取完成 750 毫秒后再读取，离开时取消读取。返回时先等待正在提交的操作，再读取最近任务。提交被拒绝时通过全局通知反馈。Go 复检已安装入口并更新对应工具的检测缓存。卡片的安装后缓存读取等待正在进行的手动刷新。Homebrew 额外验证所选前缀的目标入口。Node.js 和 npm 安装结束后并行复检两项。缺少 Node.js 时由安装 Node.js 的操作同时安装 npm，只有 Node.js 已可用时才单独提供 npm 操作。命令完成但版本未检出时，卡片显示相应结果。Homebrew、npm、uv 下载和 WinGet 维护准备阶段可取消。WinGet、Chocolatey 和 APT 进入原生安装阶段后由系统控制授权和完成。用户此时请求退出会保留窗口，并收到等待安装结束及处理可能出现的授权窗口的提示。系统强制结束会话时，退出钩子取消可中断任务并跳过原生安装器等待。页面以检测结果展示当前状态，以最近任务结果展示历史。

系统自带工具禁止卸载。开发工具页支持本机检测、安装及 WinGet 手动维护。

## 工作目录

概览页的[工作目录条](src/features/workspace/project-directory-bar.tsx) 通过 `ProjectDialog.selectDirectory()` 选择目录。原生对话框返回解析符号链接后的绝对路径，取消时返回空字符串。确认自选目录时，`ProjectDialog.validateDirectory()` 校验目录可读。概览在准备成功后提交当前选择并保存界面偏好。离开概览会取消等待，尚未提交的选择随页面丢弃。

打开概览且没有保存工作目录时，`ProjectDialog.createDefaultDirectory()` 创建或复用用户主目录下的 `.heforge-devbox`。概览通过文件夹图标打开默认目录，当前使用自选目录时提供恢复默认目录的图标，恢复前仍需预览和确认。创建默认目录和校验自选目录都在短 IPC 调用内完成。

### 保留的项目 uv 服务

项目 uv 的检测、安装、更新和卸载接口仍在 Go 服务和生成绑定中。项目操作通过 `startUVInstall(project)`、`startProjectUVUpdate(project)`、`startProjectUVUninstall(project)` 启动，由 `getUVInstall(project)` 和 `cancelUVInstall(project, id)` 读取及取消。空目录的旧 uv 安装调用仍指向同一个 Homebrew uv 任务。开发工具页撤下了项目操作入口。[早期项目工具设计](PROJECT_ENV_DESIGN.md) 记录原先的界面和流程。

[uv 发行包下载](internal/devtools/uv_release.go) 从 Astral 的 GitHub 发行版查询当前架构的安装包，查询期限为 15 秒。Grab 在五分钟内下载并核验发行包的大小和 SHA-256。Go 在应用私有临时目录中解出 uv 和 uvx，再运行新下载的 uv 读取版本。临时文件随任务清理。

Go 每 200 毫秒采样已下载字节，并通过任务状态返回下载百分比。

安装在项目工具目录中排他发布，更新通过原子交换替换旧版本，卸载先将安装目录原子移出。更新和卸载在提交前复核目录、安装记录及文件指纹。进入提交阶段后关闭取消入口，文件操作独立于任务状态锁，任务按提交与清理的实际结果反馈。清理前再次检查移出的目录及文件，发现变化或额外文件时保留待清理目录，并在任务结果中显示其路径。

`DevToolsService.detectProjectUV(project, refresh)` 读取 `.devbox/tools/uv/install.json` 并核验 uv、uvx 的文件指纹。已核验但系统或架构不匹配时返回 `incompatible`，工作目录不可用时返回 `project_unavailable`。Go 退出时等待活动项目 uv 任务完成清理。

## 代理设置

设置页通过 [use-proxy-settings.ts](src/features/settings/use-proxy-settings.ts) 管理表单，使用 [proxy-settings-card.tsx](src/features/settings/proxy-settings-card.tsx) 展示代理模式。默认系统代理；手动模式选择 `127.0.0.1:7890`、`127.0.0.1:7897` 或自定义主机与端口，HTTP 和 SOCKS5 使用同一个地址。代理地址不包含认证信息。

保存、恢复和连接检查通过浮动提示报告结果。保存和恢复完成后反馈写入结果，切回设置页时等待进行中的写入结束再读配置。[系统代理状态](src/features/settings/use-system-proxy.ts) 在系统模式显示时及窗口聚焦时读取，Go 复用 30 秒内的展示结果。系统代理刷新期间保留上次读取结果。系统配置和访问 `example.com` 的实际路由在同一个三列网格中展示。后端返回配置类型、路由状态、协议和去掉认证信息的地址。前端按这些字段合并重复项并生成中文标签。macOS 上有自动配置时单独列出实际路由。其他平台展示实际路由。

| 前端调用 | 行为 |
| --- | --- |
| `ProxyService.get()` | 返回已保存设置。文件缺失时返回默认值，文件内容无效或版本不兼容时返回 `null`，其他读取错误拒绝 |
| `ProxyService.getSystemInfo()` | 返回 30 秒内的结果，过期时读取系统配置和示例路由。取消可终止 macOS 配置查询和子进程路由查询。进程内的 MyGo 平台解析会继续到返回 |
| `ProxyService.save(settings)` | 验证并保存设置，成功后更新服务状态。非手动模式保留上次的手动地址 |
| `ProxyService.reset()` | 用系统模式和 `127.0.0.1:7890` 替换设置文件，成功后返回默认设置 |
| `ProxyService.check(settings)` | 用草稿对 `https://example.com/` 发起 HEAD 请求，成功时返回 `void`。请求使用十秒截止时间，可取消手动探测、子进程路由查询和网络请求 |

代理配置保存在 `os.UserConfigDir()/devbox/proxy.json`，当前格式版本为 `1`。设置页识别无效文件后提供恢复默认设置操作。保存或恢复失败时保留当前配置。服务首次读取或保存时加载配置，服务注册和接口生成期间无需读取。

连接检查和 uv 发行包下载共用应用内的 HTTP 代理路由。连接检查使用表单草稿，uv 发行包下载使用已保存设置。应用按同一设置为 Homebrew 与 npm 设置子进程环境，WinGet 和 APT 采用各自的系统连接设置。手动代理客户端首次请求通过 SOCKS5 握手识别所选地址，未识别到无认证 SOCKS5 时使用 HTTP，探测结果由客户端复用。应用内的系统和手动模式都绕过本机回环地址。连接检查在 HTTP `CONNECT` 阶段识别代理认证错误。

实际路由展示和连接检查共用系统代理解析。启动环境中的 `HTTP_PROXY`、`HTTPS_PROXY` 各自优先，`ALL_PROXY` 填补缺少的协议配置，`NO_PROXY` 匹配的目标直连。请求协议没有环境代理时，交由 MyGo 的 `netproxy.Proxy` 按平台解析。没有代理环境变量时，MyGo 在应用进程内按协议和主机缓存结果五秒。存在代理环境变量时，桌面端通过不继承这些变量的短命令读取系统路由，最长等待六秒。MyGo 返回空路由时直连。macOS 系统配置使用 `scutil --proxy` 读取。外部进程的代理适配见[代理设计](PROXY_DESIGN.md)。

## 修改与验证

功能修改先追踪页面、Hook、生成接口、Go Service 和具体能力的调用链。改变接口时同步处理 Go 定义、生成文件和全部调用方；改变文件格式时同步处理版本、读取规则和写入规则。

涉及职责划分、依赖方向、对外接口、状态归属、存储或生命周期的改动，需要同步更新本文。各章节记录当前实现，演进约定说明适用条件；代码、生成接口与文档保持一致。

按改动范围选择验证命令：

```sh
bun run generate
bun run build:web
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go build -o ./build/devbox .
git diff --check
```

`build:web` 包含 TypeScript 类型检查。上述检查覆盖接口生成、类型、静态检查和构建；界面交互与平台实际表现按项目规则另行验证。仅修改文档时，核对描述与源码、相对链接及 `AGENTS.md` 的文档入口。
