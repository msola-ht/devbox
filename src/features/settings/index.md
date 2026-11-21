# 设置页面索引

对应 `/settings` 路由，展示和管理 Devbox 的代理设置。

- [settings-page.tsx](settings-page.tsx)：设置页面入口。
- [proxy-settings-card.tsx](proxy-settings-card.tsx) 组合配置表单、实际路由、连接检查和配置恢复界面。
- [use-proxy-settings.ts](use-proxy-settings.ts) 管理表单选项、草稿、连接检查和跨页面写入反馈。
- [use-system-proxy.ts](use-system-proxy.ts) 在系统模式下读取代理展示结果，协调聚焦与取消。

[业务模块索引](../index.md) · [项目架构](../../../ARCHITECTURE.md)
