# 开发工具后端索引

提供开发工具检测和 macOS、Windows、Ubuntu 的本机安装。

- [detection.go](detection.go) 定义检测结果、会话内缓存、共用版本探测和带超时的命令执行。
- [discovery.go](discovery.go) 按 PATH 和工具路径发现入口，在共同期限内并行探测并合并重复结果。
- [command_paths_windows.go](command_paths_windows.go) 补查持久 PATH、WinGet 入口和 Python 安装记录。
- [command_paths_other.go](command_paths_other.go) 补查 Linux 系统工具路径。
- [source.go](source.go) 定义安装来源，提供系统命令、Apple 开发工具和 Homebrew 的路径判定。
- [git.go](git.go) 提供 Git 检测和 macOS 系统 Git 定位。
- [utilities.go](utilities.go) 提供 jq、ripgrep 和 GitHub CLI 检测。
- [powershell.go](powershell.go) 在 Windows 上检测 PowerShell 和系统自带的 Windows PowerShell。
- [zsh.go](zsh.go) 在 macOS 和 Linux 上检测 Zsh。
- [uv.go](uv.go) 提供 uv 检测与原有 uv 安装接口。
- [tool_install.go](tool_install.go) 管理安装任务的生命周期、渠道占用、输出、取消与应用退出清理。
- [install_catalog.go](install_catalog.go) 定义工具与平台适用渠道，提供本机安装接口。
- [package_install.go](package_install.go) 复检本机工具并执行 WinGet、APT 和 npm 安装。
- [windows_package_managers.go](windows_package_managers.go) 检测 WinGet 和 Chocolatey，定位系统 Windows PowerShell。
- [winget_maintenance.go](winget_maintenance.go) 使用微软模块安装、修复和更新当前用户的 WinGet。
- [chocolatey_install.go](chocolatey_install.go) 通过原生授权运行 Chocolatey 官方安装程序。
- [user_uv_install.go](user_uv_install.go) 将 Linux uv 发行包中的工具发布到用户目录。
- [project_uv.go](project_uv.go) 检测项目内的 uv 安装记录和文件。
- [project_uv_darwin.go](project_uv_darwin.go) 管理 macOS 项目内 uv 的安装、更新和卸载。
- [project_uv_other.go](project_uv_other.go) 定义其他平台的项目 uv 安装结果。
- [uv_release.go](uv_release.go) 查询、下载、核验并解包 macOS 和 Linux 的 uv 发行包。
- [python.go](python.go) 查找 Python 3 并读取各解释器的路径与版本。
- [python_launcher.go](python_launcher.go) 查询 Windows Python 安装管理器的已有解释器。
- [apple_tools.go](apple_tools.go) 为 Git 和 Python 定位 Apple 开发工具目录中的实际命令。
- [homebrew.go](homebrew.go) 检测 Homebrew，并保存安装后复检使用的路径和 Cellar 位置。
- [homebrew_install.go](homebrew_install.go) 执行 Homebrew 安装，协调同前缀任务，并复检已安装入口。
- [node.go](node.go) 检测 Node.js，为 Node.js 和 npm 查找已知版本目录并识别来源。
- [javascript.go](javascript.go) 提供 npm 和 pnpm 检测、Corepack 离线探测及 JavaScript 版本解析。
- [javascript_other.go](javascript_other.go) 在 Windows 以外直接执行 JavaScript 工具的版本命令。
- [javascript_windows.go](javascript_windows.go) 在 Windows 执行版本命令，支持 `npm.cmd` 和 `pnpm.cmd` 等批处理入口。

[后端索引](../index.md) · [前端模块](../../src/features/devtools/index.md) · [项目架构](../../ARCHITECTURE.md)
