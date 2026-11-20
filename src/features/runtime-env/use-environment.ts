import { useCallback, useEffect, useState } from "react";

import { useLatestRequest } from "@/hooks/use-latest-request";
import { EnvironmentService, type EnvironmentInfo } from "@/mygo";

interface EnvironmentState {
  info: EnvironmentInfo | null;
  refreshing: boolean;
  failed: boolean;
}

export function useEnvironment() {
  const [state, setState] = useState<EnvironmentState>({
    info: null,
    refreshing: true,
    failed: false,
  });
  const startRequest = useLatestRequest();

  const refresh = useCallback(async () => {
    const signal = startRequest();
    setState((current) => ({ ...current, refreshing: true, failed: false }));

    try {
      const info = await EnvironmentService.get.with({ signal })();
      if (!signal.aborted) {
        setState({ info, refreshing: false, failed: false });
      }
    } catch {
      if (!signal.aborted) {
        setState((current) => ({ ...current, refreshing: false, failed: true }));
      }
    }
  }, [startRequest]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { ...state, refresh };
}
