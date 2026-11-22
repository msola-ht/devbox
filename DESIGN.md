# Devbox 产品设计

本文记录 Devbox 的产品目标和后续功能设计。当前实现和接口以[架构文档](ARCHITECTURE.md)为准。

## 目标与范围

Devbox 是面向 macOS、Windows 和 Linux 的桌面开发环境工具箱，帮助用户了解本机系统信息，检查开发工具，并按需安装和管理工具。近期按工具和平台分批补齐本机安装。

当前应用提供系统信息采集、运行环境展示、本地快照、应用级代理设置、开发工具检测及按平台安装。代理设置见[全局代理设计](PROXY_DESIGN.md)。

## 运行环境

概览页展示默认登录 Shell 和用户主目录，运行环境与系统信息通过「刷新概览」统一刷新。工具检测使用 Devbox 进程的 `PATH`，并按工具补查已知的安装路径。

## 工具检测

“开发工具”页面提供本机管理清单，按系统包管理器、基础工具、语言生态及 Shell 与脚本分组。macOS 上检测 Homebrew，用于判断本机安装的前置条件。Git、GitHub CLI、jq 和 ripgrep 归入基础工具，Python、uv 和 Node.js、npm、pnpm 分别归入所属语言生态。macOS 和 Linux 检测 Zsh，Windows 检测 PowerShell。扩展时按用途归类，按工具声明平台适用范围。

工具以分组卡片展示，每项显示名称，并以首个可用结果作为摘要。全部不可用时展示首条结果。多个结果通过卡片上的数字入口打开详情，按检测顺序显示版本、路径、已识别的安装来源或失败原因。过长的路径单行省略，悬浮或聚焦时可查看完整路径。首次读取时分别检测，之后复用本次运行中的结果，并支持单项手动刷新。

| 状态 | 含义 |
| --- | --- |
| 检测中 | 正在读取该工具的信息 |
| 可用 | 已找到可执行入口，并成功获取版本 |
| 当前环境未找到 | 在查找范围内没有找到入口，或已知的 Windows Python 启动别名没有对应的已安装解释器 |
| 检测失败 | 入口无法使用、命令执行失败、版本未识别或超过检测期限 |
| 当前平台不适用 | 当前平台跳过该工具的检测 |

检测以应用进程使用的环境为基准。桌面应用与终端的 `PATH` 可能不同，各工具补充检查的安装路径见[架构文档](ARCHITECTURE.md#开发工具检测)。

Windows 上识别到 Python 应用执行别名时，通过安装管理器查询已有运行时，界面以解释器实际返回的版本与路径为准。 [Python Windows 文档](https://docs.python.org/3/using/windows.html)

Corepack 探测关闭联网，默认版本缺少缓存时提示用户安装所需版本。

检测结果存于本次 Go 进程。Go 为探测设置时间预算，返回已取得的结果与失败原因。刷新或离开页面时，前端取消旧请求。

## 本机工具安装

本机安装使用系统包管理器或工具官方安装渠道，目标可以是系统位置或当前用户的工具目录。是否需要管理员权限由安装渠道决定。开发工具页继续按现有分组展示，每项安装从对应卡片发起，卡片显示安装渠道和结果。uv 的可执行文件归本机管理。Python 虚拟环境和项目 Node.js 版本选择留待工作区设计，入口随工具展示。

系统自带工具禁止卸载。WinGet 本身支持当前用户手动安装、修复和更新。后续接入卸载时，以确认来源的单个安装为操作目标，保留系统自带版本。

### 当前实现

本机安装按平台提供渠道。安装前复检已经可用的版本，运行时在卡片展示进度、授权提示和日志，结束后刷新对应工具的版本、路径和来源。

| 平台 | 渠道和范围 |
| --- | --- |
| macOS | 通过已有 Homebrew 安装 uv、ripgrep、jq、Git、GitHub CLI、Python、Node.js 24 与随附的 npm、pnpm。Git 认可可用的 Apple 开发工具版本。Homebrew 卡片提供官方安装引导，Zsh 使用系统自带版本。 [Homebrew 安装](https://docs.brew.sh/Installation) · [Node.js 24](https://formulae.brew.sh/formula/node@24) |
| Windows | WinGet 卡片支持安装、修复和手动更新，Chocolatey 卡片支持检测和通过官方引导安装。通过 WinGet 安装 uv、ripgrep、jq、Git、GitHub CLI、CPython 3.14、Node.js LTS 与随附的 npm、pnpm 和 PowerShell 7。系统自带 Windows PowerShell 保持只读，可与 PowerShell 7 并存。 [微软 WinGet 维护](https://github.com/microsoft/winget-cli/blob/master/src/PowerShell/Help/Microsoft.WinGet.Client/Repair-WinGetPackageManager.md) · [Chocolatey 安装](https://chocolatey.org/install) |
| Ubuntu | 通过 APT 安装 ripgrep、jq、Git、GitHub CLI、Python、Node.js 与 npm、Zsh，版本由发行版仓库提供。amd64 和 arm64 上从 Astral 发行包安装 uv 到 `~/.local/bin`。pnpm 12 通过已有 npm 安装到同一用户目录，需要 Node.js 18 或更新版本。 [APT](https://documentation.ubuntu.com/server/how-to/software/package-management/) · [uv 安装](https://docs.astral.sh/uv/getting-started/installation/) · [pnpm 安装](https://pnpm.io/installation) |

安装任务由 Go 管理。启动时返回任务标识，离开页面后任务继续。相同 Homebrew 前缀或同一包管理渠道只执行一个 Devbox 安装任务。系统授权交由原生安装器处理。Homebrew、npm 和 uv 下载可请求取消，卡片提示已经发生的本机变更。WinGet、Chocolatey 与 APT 的原生安装期间请求退出时，窗口保留并提示等待安装结束。

### 首批实施与验收

**macOS 通过已有 Homebrew 安装 ripgrep** 是首个复用 uv 安装流程的批次。Homebrew 的包名为 `ripgrep`，安装后用所选 Homebrew 前缀下的 `bin/rg` 验证。 [Homebrew ripgrep](https://formulae.brew.sh/formula/ripgrep)

1. `rg` 未找到且本机安装渠道适用时，卡片提供「通过 Homebrew 安装」。已有可用版本时显示现有检测结果。已有入口损坏时保留失败原因和重试检测入口。
2. Go 启动前重新检测 `rg` 和 Homebrew。找到可用 `rg` 时直接返回检测结果。缺少 Homebrew 时提供官方安装说明，用户返回后可以重试。执行时沿用已有的代理适配和日志采集，使用固定的 formula 名称，禁止 `brew install` 顺带升级已安装的目标 formula。同一个 Homebrew 前缀只运行一个 Devbox 安装任务。若另一项安装已占用该前缀，卡片提示等待当前任务完成后重试。 [Homebrew 安装说明](https://docs.brew.sh/Installation) · [Homebrew 命令参考](https://docs.brew.sh/Manpage)
3. 任务结束后验证目标路径、版本和来源，刷新该卡片。区分已可用、命令失败、命令完成但新入口未检出，以及取消。取消后复查实际安装状态。离开页面后任务继续，返回卡片可查看原任务。
4. 以现有 uv 链路为基础，按实际复用提取 Homebrew 执行、任务状态和卡片反馈。包名、验证入口和版本读取仍由各工具定义。首批覆盖 uv 安装回归和 ripgrep 自定义 Homebrew 前缀的复检。

## 职责划分

- 前端负责工具页面、操作入口、请求状态和结果展示，优先组合项目已有的 shadcn/ui 组件。
- Go 工具业务负责查找工具、读取版本、平台差异和安装编排；外部命令使用明确的可执行入口和参数，不向前端提供任意命令执行接口。
- 前后端继续使用 MyGo 生成的类型化接口。检测采用普通请求与返回；需要跨页面运行的安装采用由 Go 持有的任务状态。
- 工具状态以本机检测结果为准。安装任务状态在应用运行期间由 Go 持有；主题和路由等界面偏好继续由前端管理。

具体模块、接口、状态和生命周期随功能实现同步写入[架构文档](ARCHITECTURE.md)，并更新相关目录索引。

[项目环境方案](PROJECT_ENV_DESIGN.md) 记录早期项目工具设计。

## 验收与后续

在对应平台验证缺失时可装、首装时已有版本会跳过、安装后能读取指定入口的版本。安装器验收包含渠道缺失、授权拒绝、网络失败、可取消阶段、页面重进、以及已安装但当前进程尚未发现入口。WinGet 的更新验收确认已有较新版本时跳过。验证遵循[项目规则](AGENTS.md)。其他工具升级、可卸载安装的管理、多版本切换及项目级配置属于后续工作。
