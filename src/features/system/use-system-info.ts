import { useEffect, useState } from "react";

import { SystemService, type SystemInfo } from "@/mygo";

type SystemInfoState =
  | { status: "loading" }
  | { status: "ready"; data: SystemInfo }
  | { status: "error" };

export function useSystemInfo() {
  const [state, setState] = useState<SystemInfoState>({ status: "loading" });

  useEffect(() => {
    const controller = new AbortController();

    void SystemService.info.with({ signal: controller.signal })()
      .then((data) => {
        if (!controller.signal.aborted) setState({ status: "ready", data });
      })
      .catch(() => {
        if (!controller.signal.aborted) setState({ status: "error" });
      });

    return () => controller.abort();
  }, []);

  return state;
}
