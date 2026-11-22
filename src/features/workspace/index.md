# 概览页面索引

对应 `/` 路由。项目工作目录位于顶部。系统信息在左、运行环境在右，概览内容较窄时纵向排列。

- [workspace-page.tsx](workspace-page.tsx) 组合[系统信息卡片](../system/system-info-card.tsx)和[运行环境卡片](../runtime-env/environment-card.tsx)，统一提供概览刷新入口。
- [project-directory-bar.tsx](project-directory-bar.tsx) 在概览内读取、准备和保存工作目录，提供原生目录选择、确认、取消，以及打开和恢复默认目录。

[业务模块索引](../index.md) · [项目架构](../../../ARCHITECTURE.md)
