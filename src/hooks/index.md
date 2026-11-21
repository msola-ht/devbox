# 通用 Hook 索引

跨业务使用的 React Hook，业务专属请求与状态位于对应的 `features` 模块。

- [use-mobile.ts](use-mobile.ts)：根据视口宽度判断移动端布局。
- [use-latest-request.ts](use-latest-request.ts) 开始新请求或组件卸载时取消上次请求。
- [use-refreshable.ts](use-refreshable.ts) 协调读取、重读缓存、手动刷新、取消及界面结果保留。

[前端索引](../index.md) · [项目架构](../../ARCHITECTURE.md)
