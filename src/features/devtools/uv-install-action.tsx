import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { CardFooter } from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Spinner } from "@/components/ui/spinner";

import { useUVInstall } from "./use-uv-install";

export default function UVInstallAction({ canInstall, project, reload }: {
  canInstall: boolean;
  project: string;
  reload: () => Promise<void>;
}) {
  const { task, active, operation, actionError, readFailed, reread, submit } = useUVInstall(project, reload);
  if (!canInstall && !task && !operation && !actionError && !readFailed) return null;

  const busy = active || operation !== null;
  const message = operation === "starting"
    ? "正在启动 uv 安装…"
    : operation === "canceling" ? "正在请求取消…" : task?.message;
  const failed = !busy && (task?.status === "failed" ||
    task?.status === "not_detected" || task?.status === "missing_homebrew");

  return (
    <CardFooter className="min-w-0 flex-col items-stretch gap-2">
      <div role="status" className="flex items-center gap-2 wrap-anywhere">
        {busy && <Spinner className="shrink-0 motion-reduce:animate-none" aria-hidden />}
        {!failed && message}
      </div>
      {failed && (
        <Alert variant="destructive">
          <AlertDescription className="wrap-anywhere">{message}</AlertDescription>
        </Alert>
      )}
      {(actionError || readFailed) && (
        <Alert variant="destructive">
          <AlertDescription className="wrap-anywhere">
            {actionError || "无法读取 uv 安装状态。"}
          </AlertDescription>
        </Alert>
      )}
      <div className="flex flex-wrap gap-2">
        {readFailed && (
          <Button type="button" size="sm" variant="outline" onClick={reread}>重试读取</Button>
        )}
        {active && task ? (
          <Button
            type="button"
            size="sm"
            variant="outline"
            disabled={task.status === "canceling" || operation !== null}
            onClick={() => void submit(task.id)}
          >
            取消安装
          </Button>
        ) : canInstall && (
          <Button
            type="button"
            size="sm"
            disabled={task === undefined || readFailed || operation !== null}
            onClick={() => void submit()}
          >
            {project ? "安装到项目" : "通过 Homebrew 安装"}
          </Button>
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
      {task?.output && (
        <Collapsible>
          <CollapsibleTrigger render={<Button type="button" variant="ghost" size="sm" />}>
            安装输出
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
