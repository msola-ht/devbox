import { useCallback, useEffect, useState } from "react";
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

  const load = useCallback(async (refresh: boolean) => {
    const signal = startRequest();
    setState((current) => ({
      ...current,
      pending: refresh ? "refreshing" : "loading",
      failed: false,
    }));

    try {
      const data = await read.with({ signal })(refresh);
      if (!signal.aborted) setState({ data, pending: null, failed: false });
    } catch {
      if (!signal.aborted) {
        setState((current) => ({ ...current, pending: null, failed: true }));
      }
    }
  }, [read, startRequest]);

  useEffect(() => {
    void load(false);
  }, [load]);

  return { ...state, refresh: () => load(true) };
}
