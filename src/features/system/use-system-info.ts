import { useCallback, useEffect, useRef, useState } from "react";

import { SystemService, type SystemSnapshot } from "@/mygo";

interface SystemInfoState {
  snapshot: SystemSnapshot | null;
  refreshing: boolean;
  error: string | null;
}

export function useSystemInfo() {
  const [state, setState] = useState<SystemInfoState>({
    snapshot: null,
    refreshing: true,
    error: null,
  });
  const activeRequest = useRef<AbortController | null>(null);

  const refresh = useCallback(async () => {
    activeRequest.current?.abort();
    const controller = new AbortController();
    activeRequest.current = controller;
    setState((current) => ({ ...current, refreshing: true, error: null }));

    try {
      const snapshot = await SystemService.refresh.with({
        signal: controller.signal,
      })();
      if (!controller.signal.aborted) {
        setState({ snapshot, refreshing: false, error: null });
      }
    } catch {
      if (!controller.signal.aborted) {
        setState((current) => ({
          ...current,
          refreshing: false,
          error: "暂时无法读取系统信息，请重试。",
        }));
      }
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    activeRequest.current = controller;

    async function load() {
      try {
        const snapshot = await SystemService.getSnapshot.with({
          signal: controller.signal,
        })();
        if (controller.signal.aborted) return;
        if (snapshot) setState((current) => ({ ...current, snapshot }));
      } catch (error) {
        // A failed cache read must not prevent collecting fresh information.
        if (!controller.signal.aborted) {
          console.warn("读取系统信息快照失败", error);
        }
      }

      if (!controller.signal.aborted) await refresh();
    }

    void load();

    return () => activeRequest.current?.abort();
  }, [refresh]);

  return { ...state, refresh };
}
