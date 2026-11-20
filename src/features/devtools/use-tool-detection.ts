import { useCallback, useEffect, useState } from "react";
import type { Cancellable } from "mygo-runtime";

import { useLatestRequest } from "@/hooks/use-latest-request";
import type { ToolDetection } from "@/mygo";

type DetectionState = ToolDetection | "detecting" | "request_failed";

export function useToolDetection(detect: Cancellable<[], ToolDetection>) {
  const [state, setState] = useState<DetectionState>("detecting");
  const startRequest = useLatestRequest();

  const refresh = useCallback(async () => {
    const signal = startRequest();
    setState("detecting");

    try {
      const result = await detect.with({ signal })();
      if (!signal.aborted) setState(result);
    } catch {
      if (!signal.aborted) setState("request_failed");
    }
  }, [detect, startRequest]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { state, refresh };
}
