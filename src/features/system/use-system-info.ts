import { useCallback, useEffect, useState } from "react";

import { useLatestRequest } from "@/hooks/use-latest-request";
import { SystemService, type SystemSnapshot } from "@/mygo";

interface SystemInfoState {
  snapshot: SystemSnapshot | null;
  pending: "loading" | "refreshing" | null;
  error: string | null;
}

export function useSystemInfo() {
  const [state, setState] = useState<SystemInfoState>({
    snapshot: null,
    pending: "loading",
    error: null,
  });
  const startRequest = useLatestRequest();

  const refresh = useCallback(async () => {
    const signal = startRequest();
    setState((current) => ({ ...current, pending: "refreshing", error: null }));

    try {
      const snapshot = await SystemService.refresh.with({ signal })();
      if (!signal.aborted) {
        setState({ snapshot, pending: null, error: null });
      }
    } catch {
      if (!signal.aborted) {
        setState((current) => ({
          ...current,
          pending: null,
          error: "暂时无法读取系统信息，请重试。",
        }));
      }
    }
  }, [startRequest]);

  useEffect(() => {
    const signal = startRequest();

    async function load() {
      try {
        const snapshot = await SystemService.getSnapshot.with({ signal })();
        if (signal.aborted) return;
        if (snapshot) {
          setState({ snapshot, pending: null, error: null });
          return;
        }
      } catch (error) {
        // A failed cache read must not prevent collecting fresh information.
        if (!signal.aborted) {
          console.warn("读取系统信息快照失败", error);
        }
      }

      if (!signal.aborted) await refresh();
    }

    void load();

  }, [refresh, startRequest]);

  return { ...state, refresh };
}
