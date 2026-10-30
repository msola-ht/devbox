import {
  CheckCircle2Icon,
  HardDriveIcon,
  PackageIcon,
  TerminalIcon,
  TriangleAlertIcon,
} from "lucide-react"

import type {
  InstallationSummary,
  ToolSummary,
} from "@/features/toolchain/contracts"
import {
  getToolStatus,
  installationKey,
  isActiveManagedInstallation,
} from "@/features/toolchain/model"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

function InstallationRow({
  installation,
  active,
}: {
  installation: InstallationSummary
  active: boolean
}) {
  const sourceLabel =
    installation.source === "managed"
      ? "DevBox 托管"
      : (installation.sourceHint ?? "系统安装")

  return (
    <div className="rounded-lg border bg-muted/30 p-3">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <TerminalIcon className="size-4 shrink-0 text-muted-foreground" />
          <span className="font-mono text-sm font-medium">
            {installation.version ?? "版本未知"}
          </span>
          {active ? (
            <span className="inline-flex items-center gap-1 rounded-full bg-primary px-2 py-0.5 text-[11px] text-primary-foreground">
              <CheckCircle2Icon className="size-3" />
              当前
            </span>
          ) : null}
        </div>
        <span className="rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
          {sourceLabel}
        </span>
      </div>
      <p className="mt-2 font-mono text-xs leading-5 break-all text-muted-foreground">
        {installation.path}
      </p>
      {!installation.healthy ? (
        <p className="mt-2 flex items-center gap-1 text-xs text-destructive">
          <TriangleAlertIcon className="size-3.5" />
          {installation.errorCode ?? "安装状态异常"}
        </p>
      ) : null}
    </div>
  )
}

export function ToolSummaryCard({ tool }: { tool: ToolSummary }) {
  const status = getToolStatus(tool)
  const allInstallations = [
    ...tool.systemInstallations,
    ...tool.managedInstallations,
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <PackageIcon className="size-4 text-muted-foreground" />
          {tool.displayName}
        </CardTitle>
        <CardDescription>{tool.description}</CardDescription>
        <CardAction>
          <span
            className={
              status.active
                ? "rounded-full bg-primary px-2.5 py-1 text-xs text-primary-foreground"
                : "rounded-full bg-muted px-2.5 py-1 text-xs text-muted-foreground"
            }
          >
            {status.label}
          </span>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        {allInstallations.length === 0 ? (
          <div className="rounded-lg border border-dashed p-6 text-center">
            <p className="text-sm font-medium">未发现 {tool.displayName}</p>
            <p className="mt-1 text-xs leading-5 text-muted-foreground">
              当前阶段只执行只读检测，安装与版本切换将在后续阶段接入。
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {allInstallations.map((installation) => (
              <InstallationRow
                key={installationKey(installation)}
                installation={installation}
                active={isActiveManagedInstallation(tool, installation)}
              />
            ))}
          </div>
        )}
      </CardContent>
      <CardFooter className="gap-2 text-xs text-muted-foreground">
        <HardDriveIcon className="size-3.5" />
        检测结果为只读，不会修改系统安装
      </CardFooter>
    </Card>
  )
}
