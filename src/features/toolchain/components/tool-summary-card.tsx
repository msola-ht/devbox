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
  getToolStatus,
  installationKey,
  isActiveManagedInstallation,
  isDefaultShellInstallation,
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
  currentLabel,
}: {
  installation: InstallationSummary
  currentLabel: string | null
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
        {currentLabel ? (
          <Badge>
            <CheckCircle2Icon data-icon="inline-start" />
            {currentLabel}
          </Badge>
        ) : null}
        <Badge variant="secondary">{sourceLabel}</Badge>
      </ItemActions>
    </Item>
  )
}

export function ToolSummaryCard({ tool }: { tool: ToolSummary }) {
  const status = getToolStatus(tool)
  const currentInstallation = tool.effectiveInstallation
  const installations =
    tool.kind === "shell"
      ? tool.shellInstallations
      : currentInstallation
        ? [currentInstallation]
        : []

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <PackageIcon className="size-4 text-muted-foreground" />
          {tool.displayName}
          {tool.isDefaultShell ? (
            <Badge variant="outline">系统 Shell 默认</Badge>
          ) : null}
        </CardTitle>
        <CardDescription>{tool.description}</CardDescription>
        {status.label ? (
          <CardAction className="flex flex-col items-end gap-1">
            <Badge variant={status.active ? "default" : "secondary"}>
              {status.label}
            </Badge>
            {tool.activeInstallationErrorCode ? (
              <Badge variant="destructive">
                {tool.activeInstallationErrorCode}
              </Badge>
            ) : null}
          </CardAction>
        ) : null}
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {installations.length === 0 ? (
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <PackageOpenIcon />
              </EmptyMedia>
              <EmptyTitle>当前环境未发现 {tool.displayName}</EmptyTitle>
              <EmptyDescription>
                默认 Shell 登录 PATH
                中没有当前生效版本。当前阶段只执行只读检测，安装与版本切换将在后续阶段接入。
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <ItemGroup className="gap-2">
            {installations.map((installation) => (
              <InstallationRow
                key={installationKey(installation)}
                installation={installation}
                currentLabel={
                  isDefaultShellInstallation(tool, installation)
                    ? "默认"
                    : isActiveManagedInstallation(tool, installation)
                      ? "当前"
                      : null
                }
              />
            ))}
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
