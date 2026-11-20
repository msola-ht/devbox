import { useCallback, useEffect, useState } from "react";

import { useLatestRequest } from "@/hooks/use-latest-request";
import { DevToolsService, type HomebrewDetection } from "@/mygo";

export function useHomebrew() {
  const [result, setResult] = useState<HomebrewDetection | null>(null);
  const startRequest = useLatestRequest();

  const refresh = useCallback(async () => {
    const signal = startRequest();
    setResult(null);

    try {
      const detection = await DevToolsService.detectHomebrew.with({ signal })();
      if (!signal.aborted) setResult(detection);
    } catch {
      if (!signal.aborted) {
        setResult({
          status: "failed",
          path: "",
          version: "",
          message: "暂时无法读取 Homebrew 信息，请重试。",
        });
      }
    }
  }, [startRequest]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { result, refresh };
}
