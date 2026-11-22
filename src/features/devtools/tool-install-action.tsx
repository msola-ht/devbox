import { useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { CardFooter } from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Spinner } from "@/components/ui/spinner";
import { Progress, ProgressLabel, ProgressValue } from "@/components/ui/progress";
import type { InstallChannel, InstallTool, ToolDetection } from "@/mygo";

import { useToolInstall } from "./use-tool-install";

type ManagedChannel = Exclude<InstallChannel, "homebrew_setup">;
const installLabels: Record<ManagedChannel, string> = {
  homebrew: "通过 Homebrew 安装",
  winget: "通过 WinGet 安装",
  app_installer: "安装 WinGet",
  chocolatey: "安装 Chocolatey",
  apt: "通过 APT 安装",
  astral: "安装到用户工具目录",
  npm: "通过 npm 安装到用户目录",
};
const installHelp: Partial<Record<ManagedChannel, { label: string; href: string }>> = {
  homebrew: { label: "Homebrew 安装说明", href: "https://docs.brew.sh/Installation" },
  winget: { label: "WinGet 安装说明", href: "https://learn.microsoft.com/windows/package-manager/winget/#install-winget" },
  app_installer: { label: "WinGet 官方说明", href: "https://learn.microsoft.com/windows/package-manager/winget/#install-winget" },
  chocolatey: { label: "Chocolatey 官方安装说明", href: "https://chocolatey.org/install" },
  apt: { label: "APT 安装说明", href: "https://documentation.ubuntu.com/server/how-to/software/package-management/" },
};

export default function ToolInstallAction({ tool, channel, name, results, ready, reload, hideInstallButton }: {
  tool: InstallTool;
  channel: ManagedChannel;
  name: string;
  results: ToolDetection[];
  ready: boolean;
  reload: () => Promise<void>;
  hideInstallButton: boolean;
}) {
  const { task, active, rechecking, operation, readFailed, reread, submit } =
    useToolInstall(tool, name, reload);
  const busy = active || operation !== null;
  const [focusedAction, setFocusedAction] = useState<"install" | "update" | null>(null);
  const powerShell7 = tool === "powershell";
  const winGetMaintenance = tool === "winget";
  const managedInstaller = winGetMaintenance || tool === "chocolatey";
  const chocolateyUnavailable = tool === "chocolatey" && results.some(({ status }) => status === "failed");
  const available = results.some((result) => result.status === "available" &&
    (!powerShell7 || result.source !== "system" && Number(result.version.split(".")[0]) >= 7));
  const canInstall = !hideInstallButton && !available && !chocolateyUnavailable &&
    results.some(({ status }) => status === "not_found" || status === "failed" ||
      (powerShell7 && status === "available"));
  const canMaintain = winGetMaintenance && available;
  const canStart = canInstall || canMaintain;
  const bundle = tool === "node" || (tool === "npm" && channel !== "apt");
  const target = powerShell7 ? "PowerShell 7" : bundle ? "Node.js 和 npm" : name;
  const repair = winGetMaintenance && (available || results.some(({ status }) => status === "failed"));
  const installLabel = repair ? "修复 WinGet"
    : powerShell7 || bundle ? `${installLabels[channel]} ${target}` : installLabels[channel];
  const missingInstaller = task?.status === "missing_homebrew" || task?.status === "missing_installer";
  const help = installHelp[channel];
  const failed = !busy && (task?.status === "failed" ||
    task?.status === "not_detected" || missingInstaller);
  const taskMessage = task && (active ? task.message
    : `${winGetMaintenance ? "上次操作结果" : "上次安装结果"}：${task.message}`);
  const message = operation === "canceling" ? "正在请求取消…"
    : operation === "update" ? "正在启动 WinGet 更新…"
    : operation === "install" ? (repair ? "正在启动 WinGet 修复…" : `正在启动 ${target} 安装…`)
    : failed ? undefined : taskMessage;
  const downloadPercent = active && task?.status === "running" ? task.downloadPercent : null;
  const canCancel = active && task?.cancelable && operation === null;
  const actionLabel = operation === "canceling" || task?.status === "canceling" ? "正在取消"
    : canCancel ? (task.action === "update" ? "取消更新" : "取消安装")
    : busy ? (winGetMaintenance ? "正在维护" : "正在安装")
    : canStart ? installLabel : "安装操作已结束";
  const startDisabled = busy || !ready || rechecking || task === undefined || readFailed;
  const actionDisabled = !canCancel && (!canStart || startDisabled);

  if (!canStart && !task && !busy && !readFailed && !chocolateyUnavailable && !focusedAction) return null;

  return (
    <CardFooter className="min-w-0 flex-col items-stretch gap-2">
      {downloadPercent != null && downloadPercent < 100 && (
        <Progress value={downloadPercent}>
          <ProgressLabel>正在下载 {target}</ProgressLabel>
          <ProgressValue />
        </Progress>
      )}
      <div role="status" className={message ? "flex items-center gap-2 wrap-anywhere" : "sr-only"}>
        {message && busy && <Spinner className="shrink-0 motion-reduce:animate-none" aria-hidden />}
        {message}
      </div>
      {failed && (
        <Alert variant="destructive">
          <AlertDescription className="wrap-anywhere">{taskMessage}</AlertDescription>
        </Alert>
      )}
      {readFailed && (
        <Alert variant="destructive">
          <AlertDescription className="wrap-anywhere">
            {active ? "无法读取安装任务状态，正在重试。" : "无法读取安装任务状态，请重试。"}
          </AlertDescription>
        </Alert>
      )}
      <div className="flex flex-wrap gap-2 empty:hidden">
        {readFailed && (
          <Button type="button" size="sm" variant="outline" disabled={operation !== null} onClick={reread}>
            立即重试
          </Button>
        )}
        {(canMaintain || focusedAction === "update" || operation === "update" ||
          (winGetMaintenance && active && task?.action === "update")) && (
          <Button type="button" size="sm" disabled={startDisabled || !canMaintain}
            focusableWhenDisabled className="aria-disabled:opacity-50"
            onFocus={() => setFocusedAction("update")}
            onBlur={() => setFocusedAction(null)}
            onClick={() => void submit("update")}>
            更新 WinGet
          </Button>
        )}
        {(busy || canStart || focusedAction === "install") && (
          <Button
            type="button"
            size="sm"
            variant={busy || canMaintain ? "outline" : "default"}
            className="h-auto min-h-7 max-w-full whitespace-normal aria-disabled:opacity-50"
            aria-label={actionLabel === installLabel && (powerShell7 || bundle || managedInstaller)
              ? installLabel : `${actionLabel} ${target}`}
            disabled={actionDisabled}
            focusableWhenDisabled
            onFocus={() => setFocusedAction("install")}
            onBlur={() => setFocusedAction(null)}
            onClick={() => void submit(canCancel ? task?.id : undefined)}
          >
            {actionLabel}
          </Button>
        )}
        {!busy && ((canStart && (missingInstaller || managedInstaller && failed)) || chocolateyUnavailable) && help && (
          <Button
            size="sm"
            variant="outline"
            nativeButton={false}
            render={<a href={help.href} target="_blank" rel="noreferrer" />}
          >
            {help.label}
          </Button>
        )}
      </div>
      {task?.output && (
        <Collapsible>
          <CollapsibleTrigger render={<Button type="button" variant="ghost" size="sm" />}>
            任务输出
          </CollapsibleTrigger>
          <CollapsibleContent>
            <pre tabIndex={0} className="max-h-36 overflow-auto whitespace-pre-wrap wrap-anywhere">
              {task.output}
            </pre>
          </CollapsibleContent>
        </Collapsible>
      )}
    </CardFooter>
  );
}
