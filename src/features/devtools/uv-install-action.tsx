import { useEffect, useRef, useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { CardFooter } from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Progress, ProgressLabel, ProgressValue } from "@/components/ui/progress";
import { Spinner } from "@/components/ui/spinner";
import type { ToolDetection } from "@/mygo";

import { useUVInstall } from "./use-uv-install";

const actionLabels = { install: "安装", update: "更新", uninstall: "卸载" };

export default function UVInstallAction({ detection, ready, project, reload }: {
  detection: ToolDetection | undefined;
  ready: boolean;
  project: string;
  reload: () => Promise<void>;
}) {
  const { task, active, rechecking, operation, actionError, readFailed, reread, submit } = useUVInstall(project, reload);
  const taskID = task?.id;
  const taskStatus = task?.status;
  const [visibleDownloadTask, setVisibleDownloadTask] = useState<number>();

  useEffect(() => {
    if (taskStatus === "running") {
      setVisibleDownloadTask(taskID);
    } else if (taskStatus === "succeeded" && taskID === visibleDownloadTask) {
      const timer = window.setTimeout(() => setVisibleDownloadTask(undefined), 1000);
      return () => window.clearTimeout(timer);
    }
  }, [taskID, taskStatus, visibleDownloadTask]);

  const busy = active || operation !== null;
  const status = detection?.status;
  const canInstall = project ? status === "not_found" : !!detection && !detection.version;
  const canUpdate = !!project && status === "available";
  const canReinstall = !!project && status === "incompatible";
  const canManage = canUpdate || canReinstall;
  const disabled = !ready || rechecking || task === undefined || readFailed || busy;
  const showCompletedDownload = taskStatus === "succeeded" && taskID === visibleDownloadTask &&
    status !== "project_unavailable" && (status === "available" || !ready || rechecking);
  const downloadPercent = operation === null && (taskStatus === "running" || showCompletedDownload)
    ? task?.downloadPercent ?? null
    : null;
  const showTask = status !== "project_unavailable" || active;
  const showSuccess = !project || (ready && !rechecking && (
    (task?.action === "uninstall" ? status === "not_found" : canUpdate)
  ));
  const message = operation === "canceling"
    ? "正在请求取消…"
    : operation ? `正在启动 uv ${actionLabels[operation]}…`
    : showTask && (task?.status !== "succeeded" || showSuccess) ? task?.message : undefined;
  const output = showTask && task?.output;
  const failed = !busy && !!message && (task?.status === "failed" ||
    task?.status === "not_detected" || task?.status === "missing_homebrew");
  const statusMessage = (downloadPercent === null || downloadPercent === 100) && !failed ? message : undefined;
  if (!canInstall && !canManage && !busy && downloadPercent === null &&
    !message && !output && !actionError && !readFailed) return null;

  return (
    <CardFooter className="min-w-0 flex-col items-stretch gap-2">
      {downloadPercent !== null && (
        <Progress value={downloadPercent}>
          <ProgressLabel>{downloadPercent === 100 ? "uv 下载完成" : "正在下载 uv…"}</ProgressLabel>
          <ProgressValue />
        </Progress>
      )}
      <div role="status" className={statusMessage ? "flex items-center gap-2 wrap-anywhere" : "sr-only"}>
        {statusMessage && busy && <Spinner className="shrink-0 motion-reduce:animate-none" aria-hidden />}
        {statusMessage}
      </div>
      {failed && (
        <Alert variant="destructive">
          <AlertDescription className="wrap-anywhere">{message}</AlertDescription>
        </Alert>
      )}
      {(actionError || readFailed) && (
        <Alert variant="destructive">
          <AlertDescription className="wrap-anywhere">
            {actionError || "无法读取 uv 任务状态，正在重试。"}
          </AlertDescription>
        </Alert>
      )}
      <div className="flex flex-wrap gap-2">
        {readFailed && (
          <Button type="button" size="sm" variant="outline" disabled={operation !== null} onClick={reread}>
            立即重试
          </Button>
        )}
        {active && task && (task.cancelable || task.status === "canceling") ? (
          <Button
            type="button"
            size="sm"
            variant="outline"
            disabled={task.status === "canceling" || operation !== null}
            onClick={() => void submit(task.id)}
          >
            {task.status === "canceling" ? "正在取消" : `取消${actionLabels[task.action]}`}
          </Button>
        ) : !active && canInstall && (
          <Button
            type="button"
            size="sm"
            disabled={disabled}
            onClick={() => void submit("install")}
          >
            {project ? "安装到项目" : "通过 Homebrew 安装"}
          </Button>
        )}
        {!busy && canManage && (
          <>
            {canReinstall ? (
              <ConfirmProjectUVAction action="reinstall" project={project} disabled={disabled}
                onConfirm={() => void submit("update")} />
            ) : (
              <Button type="button" size="sm" variant="outline" disabled={disabled}
                onClick={() => void submit("update")}>
                更新 uv
              </Button>
            )}
            <ConfirmProjectUVAction action="uninstall" project={project} disabled={disabled}
              onConfirm={() => void submit("uninstall")} />
          </>
        )}
        {task?.status === "missing_homebrew" && (
          <Button
            size="sm"
            variant="outline"
            nativeButton={false}
            render={<a href="https://brew.sh/" target="_blank" rel="noreferrer" />}
          >
            Homebrew 安装说明
          </Button>
        )}
      </div>
      {output && (
        <Collapsible>
          <CollapsibleTrigger render={<Button type="button" variant="ghost" size="sm" />}>
            任务输出
          </CollapsibleTrigger>
          <CollapsibleContent>
            <pre tabIndex={0} className="max-h-36 overflow-auto whitespace-pre-wrap wrap-anywhere">
              {output}
            </pre>
          </CollapsibleContent>
        </Collapsible>
      )}
    </CardFooter>
  );
}

function ConfirmProjectUVAction({ action, project, disabled, onConfirm }: {
  action: "reinstall" | "uninstall";
  project: string;
  disabled: boolean;
  onConfirm: () => void;
}) {
  const [open, setOpen] = useState(false);
  const cancelButton = useRef<HTMLButtonElement>(null);
  const reinstall = action === "reinstall";
  const label = reinstall ? "重新安装" : "卸载";

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger render={<Button type="button" size="sm" variant={reinstall ? "default" : "ghost"} disabled={disabled} />}>
        {label} uv
      </AlertDialogTrigger>
      <AlertDialogContent initialFocus={cancelButton}>
        <AlertDialogHeader>
          <AlertDialogTitle>{label}项目 uv？</AlertDialogTitle>
          <AlertDialogDescription className="wrap-anywhere">
            {reinstall
              ? "下载后将用适合当前架构的版本替换此工作目录下的 uv 和 uvx。"
              : "将移除此工作目录下的 uv、uvx 和安装记录。"}
            <code className="block">{project}</code>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel ref={cancelButton}>取消</AlertDialogCancel>
          <AlertDialogAction type="button" variant={reinstall ? "default" : "destructive"} disabled={disabled}
            onClick={() => {
              setOpen(false);
              onConfirm();
            }}>
            {label}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
