# 代理后端索引

提供 Devbox 的代理设置、路由解析和连接检查。

- [model.go](model.go)：配置、请求与校验。
- [store.go](store.go)：代理配置文件读写。
- [network.go](network.go)：代理选择和连接检查。
- [system_route.go](system_route.go) 在独立子进程中解析缺少环境代理的系统路由。
- [service.go](service.go) 持有配置状态，提供读取、保存、恢复和检查接口。
- [system_darwin.go](system_darwin.go)：读取 macOS 系统代理配置。
- [system_other.go](system_other.go)：其他平台系统配置展示约定。

[后端索引](../index.md) · [项目架构](../../ARCHITECTURE.md)
