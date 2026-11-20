import { useCallback, useEffect, useState } from "react";
import type { Cancellable } from "mygo-runtime";

import { useLatestRequest } from "@/hooks/use-latest-request";
import type { ToolDetection } from "@/mygo";

interface DetectionState {
  results: ToolDetection[];
  status: "detecting" | "ready" | "request_failed";
}

export function useToolDetection(detect: Cancellable<[], ToolDetection[]>) {
  const [state, setState] = useState<DetectionState>({
    results: [],
    status: "detecting",
  });
  const startRequest = useLatestRequest();

  const refresh = useCallback(async () => {
    const signal = startRequest();
    setState((current) => ({ ...current, status: "detecting" }));

    try {
      const result = await detect.with({ signal })();
      if (!signal.aborted) setState({ results: result, status: "ready" });
    } catch {
      if (!signal.aborted) {
        setState((current) => ({ ...current, status: "request_failed" }));
      }
    }
  }, [detect, startRequest]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { ...state, refresh };
}
