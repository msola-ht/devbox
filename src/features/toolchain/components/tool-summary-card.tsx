import {
  CheckCircle2Icon,
  HardDriveIcon,
  PackageOpenIcon,
  PackageIcon,
  TerminalIcon,
  TriangleAlertIcon,
} from "lucide-react"

import type {
  InstallationSummary,
  ToolSummary,
} from "@/features/toolchain/contracts"
import {
  getCurrentInstallation,
  getToolStatus,
  installationKey,
  isActiveManagedInstallation,
  isDefaultShellTool,
} from "@/features/toolchain/model"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty"
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemTitle,
} from "@/components/ui/item"

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
    <Item variant="muted" size="sm" className="items-start">
      <ItemMedia variant="icon">
        <TerminalIcon className="text-muted-foreground" />
      </ItemMedia>
      <ItemContent>
        <ItemTitle className="font-mono">
          {installation.version ?? "版本未知"}
        </ItemTitle>
        <ItemDescription className="line-clamp-none font-mono text-xs break-all">
          {installation.path}
        </ItemDescription>
        {!installation.healthy ? (
          <Badge variant="destructive">
            <TriangleAlertIcon data-icon="inline-start" />
            {installation.errorCode ?? "安装状态异常"}
          </Badge>
        ) : null}
      </ItemContent>
      <ItemActions className="flex-col items-end gap-1">
        {active ? (
          <Badge>
            <CheckCircle2Icon data-icon="inline-start" />
            当前
          </Badge>
        ) : null}
        <Badge variant="secondary">{sourceLabel}</Badge>
      </ItemActions>
    </Item>
  )
}

export function ToolSummaryCard({
  tool,
  defaultShell = null,
}: {
  tool: ToolSummary
  defaultShell?: string | null
}) {
  const status = getToolStatus(tool)
  const isDefaultShell = isDefaultShellTool(tool, defaultShell)
  const currentInstallation = getCurrentInstallation(tool)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <PackageIcon className="size-4 text-muted-foreground" />
          {tool.displayName}
          {isDefaultShell ? (
            <Badge variant="outline">系统 Shell 默认</Badge>
          ) : null}
        </CardTitle>
        <CardDescription>{tool.description}</CardDescription>
        {status.label ? (
          <CardAction>
            <Badge variant={status.active ? "default" : "secondary"}>
              {status.label}
            </Badge>
          </CardAction>
        ) : null}
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {!currentInstallation ? (
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <PackageOpenIcon />
              </EmptyMedia>
              <EmptyTitle>当前环境未发现 {tool.displayName}</EmptyTitle>
              <EmptyDescription>
                默认 Shell 登录 PATH
                中没有可用版本。当前阶段只执行只读检测，安装与版本切换将在后续阶段接入。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <ItemGroup className="gap-2">
            <InstallationRow
              key={installationKey(currentInstallation)}
              installation={currentInstallation}
              active={isActiveManagedInstallation(tool, currentInstallation)}
            />
          </ItemGroup>
        )}
      </CardContent>
      <CardFooter className="gap-2 text-xs text-muted-foreground">
        <HardDriveIcon className="size-3.5" />
        检测结果为只读，不会修改系统安装
      </CardFooter>
    </Card>
  )
}
