import { useEffect, useState } from "react";

import { ProxyService, type SystemProxyInfo } from "@/mygo";

export function useSystemProxy() {
  const [info, setInfo] = useState<SystemProxyInfo | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active: AbortController | undefined;

    async function read() {
      if (active) return;
      const request = new AbortController();
      active = request;
      setError(null);
      try {
        const result = await ProxyService.getSystemInfo.with({ signal: request.signal })();
        if (!request.signal.aborted) setInfo(result);
      } catch (err) {
        if (!request.signal.aborted) {
          setError(err instanceof Error && err.message ? err.message : "无法读取系统代理。");
        }
      } finally {
        active = undefined;
      }
    }

    void read();
    window.addEventListener("focus", read);
    return () => {
      window.removeEventListener("focus", read);
      active?.abort();
    };
  }, []);

  return { info, error };
}
