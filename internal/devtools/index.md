# 开发工具后端索引

提供 Git、uv、Python 和 macOS Homebrew 检测服务。

- [detection.go](detection.go) 定义检测结果、版本读取和带超时的命令执行。
- [discovery.go](discovery.go) 按 PATH 和工具路径发现入口，并行探测并合并重复结果。
- [git.go](git.go) 提供 Git 检测和 macOS 系统 Git 定位。
- [uv.go](uv.go) 提供 uv 检测。
- [python.go](python.go) 查找 Python 3 并读取各解释器的路径与版本。
- [python_launcher.go](python_launcher.go) 查询 Windows Python 安装管理器的已有解释器。
- [apple_tools.go](apple_tools.go) 为 Git 和 Python 定位 Apple 开发工具目录中的实际命令。
- [homebrew.go](homebrew.go) 提供 Homebrew 检测和 macOS 默认路径查找。

[后端索引](../index.md) · [前端模块](../../src/features/devtools/index.md) · [项目架构](../../ARCHITECTURE.md)
