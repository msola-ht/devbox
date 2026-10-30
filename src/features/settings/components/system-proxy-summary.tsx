import type {
  SystemProxyEndpoint,
  SystemProxyInspection,
} from "@/features/settings/contracts"
import { Badge } from "@/components/ui/badge"

export function SystemProxySummary({
  inspection,
}: {
  inspection: SystemProxyInspection
}) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <p className="text-sm font-medium">系统代理</p>
        <span className="text-xs text-muted-foreground">
          {formatCheckedAt(inspection.checkedAt)}
        </span>
      </div>
      <dl className="grid gap-x-8 gap-y-3 sm:grid-cols-2">
        <SystemProxyEndpointData label="HTTP" endpoint={inspection.http} />
        <SystemProxyEndpointData label="HTTPS" endpoint={inspection.https} />
      </dl>
    </div>
  )
}

function SystemProxyEndpointData({
  label,
  endpoint,
}: {
  label: string
  endpoint: SystemProxyEndpoint | null
}) {
  return (
    <div className="flex min-w-0 flex-col gap-1">
      <dt className="flex items-center gap-2 text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{label}</span>
        <Badge variant={endpoint ? "outline" : "secondary"}>
          {endpoint ? proxySourceLabel(endpoint.source) : "直连"}
        </Badge>
      </dt>
      {endpoint ? (
        <dd className="flex flex-col gap-1">
          <code className="text-sm break-all">{endpoint.url}</code>
          {endpoint.requiresAuthentication ? (
            <Badge variant="destructive" className="w-fit">
              需要认证
            </Badge>
          ) : null}
        </dd>
      ) : (
        <dd className="text-sm text-muted-foreground">直接连接</dd>
      )}
    </div>
  )
}

function proxySourceLabel(source: SystemProxyEndpoint["source"]): string {
  return source === "environment" ? "环境变量" : "系统设置"
}

function formatCheckedAt(value: string): string {
  const checkedAt = new Date(value)
  if (Number.isNaN(checkedAt.getTime())) {
    return "读取时间未知"
  }
  return `读取于 ${checkedAt.toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  })}`
}
