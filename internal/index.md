# 后端索引

Go 本机业务与基础模块。[main.go](../main.go) 注册系统信息、运行环境、开发工具和代理服务，前端通过 MyGo 生成接口调用。

- [systeminfo](systeminfo/index.md)：系统信息采集、快照存储与服务接口。
- [runtimeenv](runtimeenv/index.md) 提供运行环境读取。
- [devtools](devtools/index.md) 提供开发工具检测。
- [proxy](proxy/index.md) 提供代理设置、路由解析和连接检查。
- [configfile](configfile/index.md) 提供 Devbox 配置文件定位和 JSON 写入。
- [localcmd](localcmd/index.md) 运行可取消的本机命令。
- [readcache](readcache/index.md) 协调同项读取并复用会话结果。

[项目架构](../ARCHITECTURE.md) · [前端索引](../src/index.md)
