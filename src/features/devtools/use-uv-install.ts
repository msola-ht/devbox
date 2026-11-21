import { useEffect, useRef, useState } from "react";

import { DevToolsService, type UVInstallAction, type UVInstallTask } from "@/mygo";

const startTask = {
  install: DevToolsService.startUVInstall,
  update: DevToolsService.startProjectUVUpdate,
  uninstall: DevToolsService.startProjectUVUninstall,
};

function isActive(task: UVInstallTask | null | undefined) {
  return task?.status === "running" || task?.status === "canceling";
}

export function useUVInstall(project: string, reload: () => Promise<void>) {
  const [task, setTask] = useState<UVInstallTask | null>();
  const [operation, setOperation] = useState<UVInstallAction | "canceling" | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [readFailed, setReadFailed] = useState(false);
  const [revision, setRevision] = useState(0);
  const [handled, setHandled] = useState(0);
  const polling = useRef<AbortController>(null);

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
        const latest = await DevToolsService.getUVInstall.with({ signal: controller.signal })(project);
        if (controller.signal.aborted) return;
        setTask(latest);
        setReadFailed(false);
        if (isActive(latest)) {
          const downloading = latest?.downloadPercent != null && latest.downloadPercent < 100;
          schedule(downloading ? 250 : 750);
        }
      } catch {
        if (controller.signal.aborted) return;
        setReadFailed(true);
        schedule(1500);
      }
    }

    void read();
    return () => {
      controller.abort();
      window.clearTimeout(timer);
    };
  }, [operation, project, revision]);

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

  async function submit(action: UVInstallAction | number) {
    polling.current?.abort();
    setOperation(typeof action === "number" ? "canceling" : action);
    setActionError(null);
    try {
      if (typeof action === "number") {
        setTask(await DevToolsService.cancelUVInstall(project, action));
      } else {
        setTask(await startTask[action](project));
      }
    } catch (err) {
      setActionError(err instanceof Error && err.message ? err.message : "无法提交 uv 操作。");
    } finally {
      reread();
      setOperation(null);
    }
  }

  const active = isActive(task);
  const rechecking = !!task && !active && task.id !== handled;
  return { task, active, rechecking, operation, actionError, readFailed, reread, submit };
}
