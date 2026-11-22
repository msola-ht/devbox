import { useCallback, useEffect, useRef, useState } from "react";
import type { Cancellable } from "mygo-runtime";

import { useLatestRequest } from "./use-latest-request";

interface ReadState<T> {
  data: T | null;
  pending: "loading" | "refreshing" | null;
  failed: boolean;
}

export function useRefreshable<T>(read: Cancellable<[boolean], T>) {
  const [state, setState] = useState<ReadState<T>>({
    data: null,
    pending: "loading",
    failed: false,
  });
  const startRequest = useLatestRequest();
  const manualRefresh = useRef<{ signal: AbortSignal; done: Promise<boolean> } | null>(null);

  const load = useCallback((refresh: boolean) => {
    const signal = startRequest();
    setState((current) => ({
      ...current,
      pending: refresh ? "refreshing" : "loading",
      failed: false,
    }));

    const done = (async () => {
      try {
        const data = await read.with({ signal })(refresh);
        if (signal.aborted) return false;
        setState({ data, pending: null, failed: false });
        return true;
      } catch {
        if (!signal.aborted) {
          setState((current) => ({ ...current, pending: null, failed: true }));
        }
        return false;
      }
    })();
    if (refresh) {
      const request = { signal, done };
      manualRefresh.current = request;
      void done.then(() => {
        if (manualRefresh.current === request) manualRefresh.current = null;
      });
    }
    return done;
  }, [read, startRequest]);

  useEffect(() => {
    void load(false);
  }, [load]);

  const reload = useCallback(async () => {
    const refresh = manualRefresh.current;
    if (refresh) {
      const refreshed = await refresh.done;
      if (!refreshed || refresh.signal.aborted) return;
    }
    await load(false);
  }, [load]);
  return { ...state, reload, refresh: () => load(true) };
}
