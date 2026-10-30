import { LaptopIcon } from "lucide-react"

import type { EnvironmentSnapshot } from "@/features/environment/contracts"
import {
  displayGpu,
  displayOs,
  displayOsVersion,
  formatBytes,
} from "@/features/environment/model"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

type SystemInfo = Pick<
  EnvironmentSnapshot,
  | "os"
  | "osVersion"
  | "osBuild"
  | "arch"
  | "memoryTotalBytes"
  | "diskTotalBytes"
  | "gpuName"
  | "gpuCoreCount"
  | "defaultShell"
>

function EnvironmentItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex min-w-0 items-baseline justify-between gap-3">
      <dt className="shrink-0 text-xs text-muted-foreground">{label}</dt>
      <dd className="min-w-0 text-right text-sm font-medium break-words">
        {value}
      </dd>
    </div>
  )
}

export function EnvironmentSummary({
  environment,
}: {
  environment: SystemInfo
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <LaptopIcon className="size-4 text-muted-foreground" />
          系统信息
        </CardTitle>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-4 sm:grid-cols-3">
          <EnvironmentItem label="操作系统" value={displayOs(environment.os)} />
          <EnvironmentItem
            label="系统版本"
            value={displayOsVersion(environment.osVersion, environment.osBuild)}
          />
          <EnvironmentItem label="处理器架构" value={environment.arch} />
          <EnvironmentItem
            label="内存"
            value={formatBytes(environment.memoryTotalBytes)}
          />
          <EnvironmentItem
            label="硬盘"
            value={formatBytes(environment.diskTotalBytes)}
          />
          <EnvironmentItem
            label="GPU"
            value={displayGpu(environment.gpuName, environment.gpuCoreCount)}
          />
          <EnvironmentItem
            label="默认 Shell"
            value={environment.defaultShell ?? "未知"}
          />
        </dl>
      </CardContent>
    </Card>
  )
}
