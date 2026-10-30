import { FolderCogIcon, TriangleAlertIcon } from "lucide-react"

import { useToolchainWorkspace } from "@/app"
import { Page } from "@/components/page"
import {
  QueryErrorCard,
  QueryErrorNotice,
  QueryLoadingCard,
} from "@/components/query-state-card"
import { RefreshButton } from "@/components/refresh-button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { ToolchainEmptyState, ToolSummaryCard } from "@/features/toolchain"

export function ToolsPage() {
  const { environment, toolchain, loading, refreshAll } =
    useToolchainWorkspace()
  const devboxUnavailable = environment.state.data
    ? !environment.state.data.devboxReady
    : environment.state.error !== null
  const devboxStatus = environment.state.data
    ? environment.state.data.devboxReady
      ? "可写"
      : "目录异常"
    : environment.state.error
      ? "不可用"
      : "检查中"

  return (
    <Page>
      <section className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="space-y-2">
          <p className="text-sm font-medium text-muted-foreground">工具</p>
          <h1 className="font-heading text-3xl font-semibold tracking-tight">
            工具管理
          </h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
            当前提供已注册工具的本机检测。系统安装与 DevBox
            托管安装分开展示，安装和切换能力将在下一阶段接入。
          </p>
        </div>
        <RefreshButton loading={loading} onClick={() => void refreshAll()} />
      </section>

      <Card size="sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FolderCogIcon className="size-4 text-muted-foreground" />
            DevBox 托管目录
          </CardTitle>
          <CardDescription>
            DevBox 管理的工具、缓存和状态都保存在这里
          </CardDescription>
          <CardAction>
            <span
              className={
                devboxUnavailable
                  ? "rounded-full bg-destructive/10 px-2.5 py-1 text-xs text-destructive"
                  : "rounded-full bg-muted px-2.5 py-1 text-xs text-muted-foreground"
              }
            >
              {devboxStatus}
            </span>
          </CardAction>
        </CardHeader>
        <CardContent>
          {environment.state.data ? (
            <div className="flex flex-col gap-1">
              <p className="font-mono text-xs leading-5 break-all">
                {environment.state.data.devboxRoot}
              </p>
              {environment.state.data.devboxError ? (
                <p className="flex items-start gap-1 text-xs break-words text-destructive">
                  <TriangleAlertIcon className="size-3.5 shrink-0" />
                  {environment.state.data.devboxError}
                </p>
              ) : null}
            </div>
          ) : environment.state.error ? (
            <p className="flex items-center gap-1 text-xs text-destructive">
              <TriangleAlertIcon className="size-3.5" />
              {environment.state.error.code} · {environment.state.error.message}
            </p>
          ) : (
            <Skeleton className="h-4 w-2/3" />
          )}
        </CardContent>
      </Card>

      {environment.state.error && environment.state.data ? (
        <QueryErrorNotice
          code={environment.state.error.code}
          message={environment.state.error.message}
          suffix="当前显示上次成功状态"
        />
      ) : null}

      {toolchain.state.error && toolchain.state.data ? (
        <QueryErrorNotice
          code={toolchain.state.error.code}
          message={toolchain.state.error.message}
        />
      ) : null}

      {toolchain.state.data ? (
        toolchain.state.data.tools.length > 0 ? (
          <div className="grid gap-4 lg:grid-cols-2">
            {toolchain.state.data.tools.map((tool) => (
              <ToolSummaryCard key={tool.toolId} tool={tool} />
            ))}
          </div>
        ) : (
          <ToolchainEmptyState />
        )
      ) : toolchain.state.error ? (
        <QueryErrorCard
          title="无法读取工具状态"
          code={toolchain.state.error.code}
          message={toolchain.state.error.message}
          onRetry={() => void toolchain.refresh()}
        />
      ) : (
        <QueryLoadingCard title="工具状态" />
      )}
    </Page>
  )
}

export default ToolsPage
