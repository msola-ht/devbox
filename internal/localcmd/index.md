# 本机命令索引

运行带上下文的本机命令，捕获输出或交给调用方实时处理，并在取消时终止命令的子进程。

- [output.go](output.go) 组装命令、环境、工作目录和等待期限。
- [output_unix.go](output_unix.go) 使用独立进程组。
- [output_windows.go](output_windows.go) 使用 Job Object 和无控制台窗口模式。
- [output_other.go](output_other.go) 保留其他构建目标的标准库执行语义。

[后端索引](../index.md) · [项目架构](../../ARCHITECTURE.md)
