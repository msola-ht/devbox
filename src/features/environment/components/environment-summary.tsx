import { LaptopIcon, ShieldCheckIcon, TriangleAlertIcon } from "lucide-react"

import type { EnvironmentSnapshot } from "@/features/environment/contracts"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

function displayOs(os: string) {
  if (os === "macos") {
    return "macOS"
  }
  if (os === "windows") {
    return "Windows"
  }
  if (os === "linux") {
    return "Linux"
  }
  return os
}

function EnvironmentItem({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div className="min-w-0 space-y-1">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd
        className={
          mono
            ? "font-mono text-xs leading-5 break-all text-foreground"
            : "text-sm font-medium"
        }
      >
        {value}
      </dd>
    </div>
  )
}

export function EnvironmentSummary({
  environment,
}: {
  environment: EnvironmentSnapshot
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <LaptopIcon className="size-4 text-muted-foreground" />
          本机环境
        </CardTitle>
        <CardDescription>由 Rust 读取当前设备环境信息</CardDescription>
        <CardAction>
          <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2.5 py-1 text-xs text-muted-foreground">
            {environment.devboxReady ? (
              <ShieldCheckIcon className="size-3.5" />
            ) : (
              <TriangleAlertIcon className="size-3.5 text-destructive" />
            )}
            {environment.devboxReady ? "目录就绪" : "目录异常"}
          </span>
        </CardAction>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-4 sm:grid-cols-2">
          <EnvironmentItem label="操作系统" value={displayOs(environment.os)} />
          <EnvironmentItem label="处理器架构" value={environment.arch} />
          <EnvironmentItem label="用户目录" value={environment.homeDir} mono />
          <EnvironmentItem
            label="DevBox 根目录"
            value={environment.devboxRoot}
            mono
          />
        </dl>
      </CardContent>
      {environment.devboxError ? (
        <CardFooter className="text-xs text-destructive">
          {environment.devboxError}
        </CardFooter>
      ) : null}
    </Card>
  )
}
