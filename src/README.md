# 前端模块说明

[文档索引](../docs/README.md) · [架构设计](../docs/architecture.md) · [前端规则](AGENTS.md) · [后端模块说明](../src-tauri/README.md) · [开发指南](../docs/development.md)

本文描述当前前端实现。当前功能范围包括环境查询、工具状态查询与刷新、下载代理设置和连接测试；安装、版本目录、任务进度和业务事件按开发指南的实施顺序进入前端模块。

## 应用入口与连接关系

```text
main.tsx
  ThemeProvider
    RouterProvider
      AppLayout
        AppSidebar + Breadcrumb + ThemeToggle
        Outlet
          DashboardPage / ToolsPage / SettingsPage / NotFoundPage
            Page
              app coordination (cross-feature)
              feature components
                feature hooks
                  API adapter
                    lib/tauri-ipc.ts
                      Tauri IPC
```

| 文件 | 当前职责 |
| --- | --- |
| [main.tsx](main.tsx) | React 挂载、全局样式、StrictMode、ThemeProvider 和 RouterProvider |
| [router.tsx](router.tsx) | Hash Router；`/`、`/tools`、`/settings` 和未知路径统一作为 `AppLayout` 子路由 |
| [app-layout.tsx](components/app-layout.tsx) | 唯一应用壳层，组合侧栏、Breadcrumb、主题切换和内部滚动内容区 |
| [page.tsx](components/page.tsx) | 页面统一纵向容器、宽度和间距 |
| [app-sidebar.tsx](components/app-sidebar.tsx) | DevBox 品牌、工作区导航和本地设备状态占位 |
| [app-navigation.tsx](components/app-navigation.tsx) | 概览、工具、设置的真实 `NavLink` 导航 |
| [app/use-toolchain-workspace.ts](app/use-toolchain-workspace.ts) | 组合环境与工具链模块，统一联合加载状态和刷新操作 |
| [query-state-card.tsx](components/query-state-card.tsx) | 业务无关的查询加载、失败和旧数据提示 |
| [refresh-button.tsx](components/refresh-button.tsx) | 业务无关的加载态刷新按钮 |

`ThemeProvider` 在路由外层生效；页面不重复包裹 `AppLayout`、`SidebarProvider` 或侧栏。

## 路由与页面

| 页面 | 路由 | 当前内容 |
| --- | --- | --- |
| [dashboard.tsx](pages/dashboard.tsx) | `#/` | 本机环境、已注册工具、加载/失败/空注册表状态和刷新 |
| [tools.tsx](pages/tools.tsx) | `#/tools` | 托管目录状态、工具安装列表和刷新 |
| [settings.tsx](pages/settings.tsx) | `#/settings` | 下载代理模式、系统代理读取结果、地址、绕过列表、保存和连接测试 |
| [not-found.tsx](pages/not-found.tsx) | 未匹配路径 | 404 和返回概览 |

页面通过应用组合层协调多个业务模块，只组合公开出口和基础 UI 组件，不拼接 Tauri command 名。

## 业务模块

### `features/environment`

| 文件 | 职责 |
| --- | --- |
| `contracts.ts` | `EnvironmentSnapshot`，与 Rust DTO 字段一一对应 |
| `api.ts` | 调用 `get_environment` 的模块适配 |
| `hooks/use-environment.ts` | 创建模块级快照 store |
| `model.ts` | 系统信息的纯展示转换与容量格式化 |
| `components/environment-summary.tsx` | 展示 OS 版本、架构、内存、主目录所在卷容量和 GPU |
| `index.ts` | 模块公开出口 |

### `features/toolchain`

| 文件 | 职责 |
| --- | --- |
| `contracts.ts` | `ToolListSnapshot`、`ToolSummary` 和 `InstallationSummary` |
| `api.ts` | 初次查询使用 `list_tools`，显式刷新使用 `refresh_tools` |
| `hooks/use-toolchain.ts` | 创建模块级快照 store |
| `model.ts` | 派生工具总体状态、active 安装判断和安装项展示键 |
| `components/tool-summary-card.tsx` | 展示插件返回的说明、系统/托管安装、健康状态和 active 标记 |
| `components/toolchain-empty-state.tsx` | 展示成功返回但没有注册工具时的空状态 |
| `index.ts` | 模块公开出口 |

工具名称、说明和契约版本来自 Rust 插件 DTO；前端没有针对 uv 或 Python 的检测分支。新增工具检测只需在 Rust 注册表接入插件，通用工具列表自动展示。

### `features/settings`

| 文件 | 职责 |
| --- | --- |
| `contracts.ts` | `SettingsSnapshot`、系统代理读取结果、`ProxySettings` 和代理测试结果 |
| `api.ts` | 调用 `get_settings`、`update_settings` 和 `test_proxy` |
| `hooks/use-settings.ts` | 管理设置快照，保存成功后应用 Rust 返回的新 revision |
| `hooks/use-proxy-settings-form.ts` | 管理代理编辑草稿、校验、保存、测试和过期响应 |
| `model.ts` | 代理表单草稿与 IPC 配置之间的转换和即时校验 |
| `components/proxy-settings-card.tsx` | 组合模式选择、手动代理字段、反馈和操作按钮 |
| `components/system-proxy-summary.tsx` | 展示系统代理 HTTP/HTTPS 路由、来源、读取时间和认证需求 |
| `index.ts` | 模块公开出口 |

代理密码不在本阶段范围内。代理配置由 Rust 校验并写入 `.devbox/config.json`，前端不保存业务配置副本。系统代理结果只展示脱敏后的实际 HTTP/HTTPS 路由、来源和认证需求。

## 通信与快照状态

- [tauri-ipc.ts](lib/tauri-ipc.ts) 是所有业务模块共用的传输适配，只负责 Tauri invoke、桌面环境判断和错误标准化。
- 非 Tauri 页面明确返回 `DESKTOP_UNAVAILABLE`，不使用模拟数据或空列表代替本机状态。
- [use-snapshot-store.ts](hooks/use-snapshot-store.ts) 管理 `idle`、`loading`、`ready`、`error`；同一 store 的普通并发加载合并为一次请求，查询进行中的显式刷新会在当前请求完成后排队执行。
- 快照记录 `bootId` 和 `revision`。`bootId` 改变时重置 revision；较旧 revision 不覆盖已应用的新快照。
- 刷新时保留已有数据；刷新失败后状态为 `error` 且旧数据仍可展示，页面同时显示错误提示。
- 初次进入使用查询命令；用户点击刷新时才调用重检测命令。设置保存成功后直接应用 Rust 返回的新快照。
- event、channel、任务订阅和持久化业务缓存归后续操作链路。

## 基础组件与遗留文件

`components/ui` 是 shadcn/ui CLI 生成的基础组件；业务层只使用公开 API，不修改生成件内部实现。

`components/nav-main.tsx`、`nav-projects.tsx`、`nav-user.tsx` 和 `team-switcher.tsx` 作为 UI 资源保留，不参与当前应用组合树。

## 交付边界

- 前端展示 Rust 返回的环境和检测快照，业务持久数据归 Rust。
- 浏览器验证页面状态，桌面运行验证 Tauri IPC、目录初始化和进程探测。
- 下载、安装、版本目录、任务进度、取消、激活和卸载通过后续 `operations` feature 接入；代理设置模块提供网络配置。
