import { useEffect, useRef, useState } from "react";

import { DevToolsService, type UVInstallTask } from "@/mygo";

function isActive(task: UVInstallTask | null | undefined) {
  return task?.status === "running" || task?.status === "canceling";
}

export function useUVInstall(project: string, reload: () => Promise<void>) {
  const [task, setTask] = useState<UVInstallTask | null>();
  const [operation, setOperation] = useState<"starting" | "canceling" | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [readFailed, setReadFailed] = useState(false);
  const [revision, setRevision] = useState(0);
  const handled = useRef(0);

  useEffect(() => {
    if (operation) return;
    const controller = new AbortController();
    let timer: number | undefined;

    async function read() {
      try {
        const latest = await DevToolsService.getUVInstall.with({ signal: controller.signal })(project);
        if (controller.signal.aborted) return;
        setTask(latest);
        setReadFailed(false);
        if (isActive(latest)) timer = window.setTimeout(() => void read(), 750);
      } catch {
        if (!controller.signal.aborted) setReadFailed(true);
      }
    }

    void read();
    return () => {
      controller.abort();
      window.clearTimeout(timer);
    };
  }, [operation, project, revision]);

  useEffect(() => {
    if (task && !isActive(task) && task.id !== handled.current) {
      handled.current = task.id;
      void reload();
    }
  }, [task, reload]);

  const reread = () => setRevision((value) => value + 1);

  async function submit(id?: number) {
    setOperation(id === undefined ? "starting" : "canceling");
    setActionError(null);
    try {
      if (id === undefined) {
        setTask(await DevToolsService.startUVInstall(project));
      } else {
        await DevToolsService.cancelUVInstall(project, id);
      }
    } catch (err) {
      setActionError(err instanceof Error && err.message ? err.message : "无法更新 uv 安装任务。");
    } finally {
      reread();
      setOperation(null);
    }
  }

  return { task, active: isActive(task), operation, actionError, readFailed, reread, submit };
}
