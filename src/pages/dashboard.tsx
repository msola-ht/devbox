import { useToolchainWorkspace } from "@/app"
import { Page } from "@/components/page"
import {
  QueryErrorCard,
  QueryErrorNotice,
  QueryLoadingCard,
} from "@/components/query-state-card"
import { RefreshButton } from "@/components/refresh-button"
import {
  EnvironmentSummary,
  ShellEnvironmentNotice,
} from "@/features/environment"
import { ToolchainEmptyState, ToolSummaryCard } from "@/features/toolchain"

export function DashboardPage() {
  const {
    environment,
    toolchain,
    tools,
    hasCurrentToolSnapshot,
    loading,
    refreshAll,
  } = useToolchainWorkspace()

  return (
    <Page>
      <section className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="flex flex-col gap-2">
          <h1 className="font-heading text-3xl font-semibold tracking-tight">
            本机工具链
          </h1>
          <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
            检测当前系统、托管目录和已安装工具。当前阶段只读取本机状态，不会下载或修改工具。
          </p>
        </div>
        <RefreshButton loading={loading} onClick={() => void refreshAll()} />
      </section>

      {environment.state.data ? (
        <>
          <EnvironmentSummary environment={environment.state.data} />
          <ShellEnvironmentNotice environment={environment.state.data} />
          {environment.state.error ? (
            <QueryErrorNotice
              code={environment.state.error.code}
              message={environment.state.error.message}
              suffix="当前显示上次成功状态"
            />
          ) : null}
        </>
      ) : environment.state.error ? (
        <QueryErrorCard
          title="无法读取本机环境"
          code={environment.state.error.code}
          message={environment.state.error.message}
          onRetry={() => void refreshAll()}
        />
      ) : (
        <QueryLoadingCard title="本机环境" />
      )}

      <section className="flex flex-col gap-4">
        <div className="flex flex-wrap items-end justify-between gap-2">
          <div>
            <h2 className="font-heading text-lg font-semibold">工具状态</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              只展示默认 Shell 登录 PATH 中当前生效的版本
            </p>
          </div>
          {hasCurrentToolSnapshot ? (
            <span className="text-xs text-muted-foreground">
              已检测 {tools.length} 个工具
            </span>
          ) : null}
        </div>

        {toolchain.state.error && toolchain.state.data ? (
          <QueryErrorNotice
            code={toolchain.state.error.code}
            message={toolchain.state.error.message}
          />
        ) : null}

        {hasCurrentToolSnapshot ? (
          tools.length > 0 ? (
            <div className="grid gap-4 lg:grid-cols-2">
              {tools.map((tool) => (
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
            onRetry={() => void refreshAll()}
          />
        ) : (
          <QueryLoadingCard title="工具状态" />
        )}
      </section>
    </Page>
  )
}

export default DashboardPage
