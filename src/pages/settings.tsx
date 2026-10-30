import { ProxySettingsCard, useSettings } from "@/features/settings"
import { Page } from "@/components/page"
import {
  QueryErrorCard,
  QueryErrorNotice,
  QueryLoadingCard,
} from "@/components/query-state-card"

export function SettingsPage() {
  const settings = useSettings()

  return (
    <Page>
      <header className="flex flex-col gap-2">
        <h1 className="font-heading text-3xl font-semibold tracking-tight">
          网络与下载
        </h1>
        <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
          配置版本查询和工具下载使用的代理。
        </p>
      </header>

      {settings.state.data ? (
        <>
          <ProxySettingsCard
            snapshot={settings.state.data}
            onSave={settings.update}
            onTest={settings.testProxy}
          />
          {settings.state.error ? (
            <QueryErrorNotice
              code={settings.state.error.code}
              message={settings.state.error.message}
              suffix="当前显示上次成功加载的设置"
            />
          ) : null}
        </>
      ) : settings.state.error ? (
        <QueryErrorCard
          title="无法读取网络设置"
          code={settings.state.error.code}
          message={settings.state.error.message}
          onRetry={() => void settings.refresh()}
        />
      ) : (
        <QueryLoadingCard title="网络设置" />
      )}
    </Page>
  )
}

export default SettingsPage
