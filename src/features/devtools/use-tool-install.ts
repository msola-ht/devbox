import { useEffect, useRef, useState } from "react";

import { toast } from "@/components/ui/toast";
import { DevToolsService, type InstallTool, type ToolInstallTask } from "@/mygo";

// Keep a submission visible to the next card if the user navigates away during IPC.
const submissions = new Map<InstallTool, Promise<ToolInstallTask | null>>();

function isActive(task: ToolInstallTask | null | undefined) {
  return task?.status === "running" || task?.status === "canceling";
}

export function useToolInstall(tool: InstallTool, name: string, reload: () => Promise<void>) {
  const [task, setTask] = useState<ToolInstallTask | null>();
  const [operation, setOperation] = useState<"install" | "update" | "canceling" | null>(null);
  const [readFailed, setReadFailed] = useState(false);
  const [revision, setRevision] = useState(0);
  const [handled, setHandled] = useState(0);
  const polling = useRef<AbortController>(null);
  const lastKnownActive = useRef(false);

  useEffect(() => {
    if (operation) return;
    const controller = new AbortController();
    polling.current = controller;
    let timer: number | undefined;

    function schedule(delay: number) {
      timer = window.setTimeout(() => void read(), delay);
    }

    async function read() {
      if (controller.signal.aborted) return;
      try {
        await submissions.get(tool)?.catch(() => {});
        if (controller.signal.aborted) return;
        const latest = await DevToolsService.getToolInstall.with({ signal: controller.signal })(tool);
        if (controller.signal.aborted) return;
        lastKnownActive.current = isActive(latest);
        setTask(latest);
        setReadFailed(false);
        if (lastKnownActive.current) schedule(750);
      } catch {
        if (controller.signal.aborted) return;
        setReadFailed(true);
        if (lastKnownActive.current) schedule(1500);
      }
    }

    void read();
    return () => {
      controller.abort();
      window.clearTimeout(timer);
    };
  }, [operation, tool, revision]);

  useEffect(() => {
    if (task && !isActive(task) && task.id !== handled) {
      setHandled(task.id);
      void reload();
    }
  }, [task, reload, handled]);

  function reread() {
    polling.current?.abort();
    setRevision((value) => value + 1);
  }

  async function submit(action?: number | "update") {
    if (submissions.has(tool)) return;
    polling.current?.abort();
    setOperation(typeof action === "number" ? "canceling" : action ?? "install");
    try {
      const request = typeof action === "number"
        ? DevToolsService.cancelToolInstall(tool, action)
        : action === "update"
          ? DevToolsService.startWinGetUpdate()
          : DevToolsService.startToolInstall(tool);
      submissions.set(tool, request);
      const latest = await request;
      lastKnownActive.current = isActive(latest);
      setTask(latest);
    } catch (err) {
      const title = typeof action === "number" ? `无法取消 ${name} 操作`
        : action === "update" ? `无法启动 ${name} 更新` : `无法启动 ${name} 安装`;
      toast.add({ title, description: err instanceof Error ? err.message : undefined, type: "error" });
    } finally {
      submissions.delete(tool);
      reread();
      setOperation(null);
    }
  }

  const active = isActive(task);
  const rechecking = !!task && !active && task.id !== handled;
  return { task, active, rechecking, operation, readFailed, reread, submit };
}
