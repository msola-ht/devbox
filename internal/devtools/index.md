# 开发工具后端索引

提供基础工具、Python、JavaScript、Shell 和 macOS Homebrew 检测服务。

- [detection.go](detection.go) 定义检测结果、会话内结果、版本读取和带超时的命令执行。
- [discovery.go](discovery.go) 按 PATH 和工具路径发现入口，在共同期限内并行探测并合并重复结果。
- [source.go](source.go) 定义安装来源，提供系统命令、Apple 开发工具和 Homebrew 的路径判定。
- [git.go](git.go) 提供 Git 检测和 macOS 系统 Git 定位。
- [utilities.go](utilities.go) 提供 jq、ripgrep 和 GitHub CLI 检测。
- [powershell.go](powershell.go) 在 Windows 上检测 PowerShell 和系统自带的 Windows PowerShell。
- [zsh.go](zsh.go) 在 macOS 和 Linux 上检测 Zsh。
- [uv.go](uv.go) 提供 uv 检测。
- [python.go](python.go) 查找 Python 3 并读取各解释器的路径与版本。
- [python_launcher.go](python_launcher.go) 查询 Windows Python 安装管理器的已有解释器。
- [apple_tools.go](apple_tools.go) 为 Git 和 Python 定位 Apple 开发工具目录中的实际命令。
- [homebrew.go](homebrew.go) 提供 Homebrew 检测和 macOS 默认路径查找。
- [node.go](node.go) 检测 Node.js，为 Node.js 和 npm 查找已知版本目录并识别来源。
- [javascript.go](javascript.go) 提供 npm 和 pnpm 检测及共用的版本读取逻辑。
- [javascript_other.go](javascript_other.go) 在 Windows 以外直接执行 JavaScript 工具的版本命令。
- [javascript_windows.go](javascript_windows.go) 在 Windows 执行版本命令，支持 `npm.cmd` 和 `pnpm.cmd` 等批处理入口。

[后端索引](../index.md) · [前端模块](../../src/features/devtools/index.md) · [项目架构](../../ARCHITECTURE.md)
