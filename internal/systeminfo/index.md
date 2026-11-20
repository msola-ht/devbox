# 系统信息后端索引

提供 `SystemService.GetSnapshot` 与 `SystemService.Refresh`，管理本机信息采集和快照。

- [model.go](model.go)：系统信息、GPU 与采集快照的数据结构。
- [service.go](service.go) 管理服务接口、采集超时和快照的内存与文件保存。
- [collector.go](collector.go)：系统信息采集与 GPU、磁盘并行查询。
- [powershell_windows.go](powershell_windows.go) 为 Windows GPU 和磁盘查询提供 PowerShell JSON 输出。
- [snapshot_store.go](snapshot_store.go)：快照文件位置、格式校验和读写。
- [graphics_darwin.go](graphics_darwin.go)：macOS GPU 信息。
- [graphics_linux.go](graphics_linux.go)：Linux GPU 信息。
- [graphics_windows.go](graphics_windows.go)：Windows GPU 信息。
- [graphics_other.go](graphics_other.go)：其他平台的 GPU 信息返回约定。
- [storage_darwin.go](storage_darwin.go)：macOS 物理磁盘容量。
- [storage_linux.go](storage_linux.go)：Linux 物理磁盘容量。
- [storage_windows.go](storage_windows.go)：Windows 物理磁盘容量。
- [storage_other.go](storage_other.go)：其他平台的磁盘容量返回约定。

[后端索引](../index.md) · [前端模块](../../src/features/system/index.md) · [项目架构](../../ARCHITECTURE.md)
