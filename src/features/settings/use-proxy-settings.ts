import { useCallback, useEffect, useRef, useState } from "react";

import { toast } from "@/components/ui/toast";
import { ProxyService, type Settings } from "@/mygo";

export const proxyModes = [
  { value: "system", label: "系统代理" },
  { value: "manual", label: "手动代理" },
  { value: "direct", label: "直连" },
] as const;

export const addressChoices = [
  { value: "127.0.0.1:7890", label: "127.0.0.1:7890" },
  { value: "127.0.0.1:7897", label: "127.0.0.1:7897" },
  { value: "custom", label: "自定义" },
] as const;

interface ProxyDraft {
  mode: (typeof proxyModes)[number]["value"];
  choice: (typeof addressChoices)[number]["value"];
  customAddress: string;
}

type SettingsState =
  | { status: "loading" }
  | { status: "invalid" }
  | { status: "failed"; error: string }
  | { status: "ready"; draft: ProxyDraft };

type Action = "saving" | "resetting" | "checking";

// Writes belong to the application, so a returning settings page reads after them.
let pendingWrite: Promise<Settings> | null = null;

function draftFrom(settings: Settings, keepCustom = false): ProxyDraft {
  const address = settings.manual.address;
  const preset = addressChoices.find(({ value }) => value === address);
  const choice = keepCustom ? "custom" : preset?.value ?? "custom";
  return {
    mode: settings.mode as ProxyDraft["mode"],
    choice,
    customAddress: choice === "custom" ? address : "",
  };
}

function settingsFrom(draft: ProxyDraft): Settings {
  const address = draft.choice === "custom" ? draft.customAddress : draft.choice;
  return { mode: draft.mode, manual: { address } };
}

function messageFrom(error: unknown, fallback: string) {
  return error instanceof Error && error.message ? error.message : fallback;
}

export function useProxySettings() {
  const [settings, setSettings] = useState<SettingsState>({ status: "loading" });
  const [action, setAction] = useState<Action | null>(null);
  const activeRequest = useRef<AbortController | Promise<Settings> | null>(null);
  const draft = settings.status === "ready" ? settings.draft : null;

  const load = useCallback(async () => {
    if (activeRequest.current) return;
    const controller = new AbortController();
    activeRequest.current = controller;
    setSettings({ status: "loading" });
    try {
      if (pendingWrite) await pendingWrite.catch(() => undefined);
      if (controller.signal.aborted) return;
      const result = await ProxyService.get.with({ signal: controller.signal })();
      if (!controller.signal.aborted) {
        setSettings(result ? { status: "ready", draft: draftFrom(result) } : { status: "invalid" });
      }
    } catch (err) {
      if (!controller.signal.aborted) {
        setSettings({ status: "failed", error: messageFrom(err, "无法读取代理设置。") });
      }
    } finally {
      if (activeRequest.current === controller) activeRequest.current = null;
    }
  }, []);

  useEffect(() => {
    void load();
    return () => {
      const request = activeRequest.current;
      if (request instanceof AbortController) request.abort();
      activeRequest.current = null;
    };
  }, [load]);

  function change(changes: Partial<ProxyDraft>) {
    setSettings((current) => current.status === "ready"
      ? { status: "ready", draft: { ...current.draft, ...changes } }
      : current);
  }

  async function writeSettings(next: ProxyDraft | null) {
    if (activeRequest.current || pendingWrite) return;
    const request = next ? ProxyService.save(settingsFrom(next)) : ProxyService.reset();
    activeRequest.current = request;
    pendingWrite = request;
    setAction(next ? "saving" : "resetting");
    try {
      const result = await request;
      if (activeRequest.current === request) {
        const keepCustom = next?.mode === "manual" && next.choice === "custom";
        setSettings({ status: "ready", draft: draftFrom(result, keepCustom) });
      }
      toast.add({ title: next ? "设置已保存" : "已恢复默认设置", type: "success" });
    } catch (err) {
      const title = next ? "保存失败" : "恢复失败";
      toast.add({ title, description: messageFrom(err, title), type: "error" });
    } finally {
      pendingWrite = null;
      if (activeRequest.current === request) {
        activeRequest.current = null;
        setAction(null);
      }
    }
  }

  async function check() {
    if (!draft || activeRequest.current) return;
    const controller = new AbortController();
    activeRequest.current = controller;
    setAction("checking");
    try {
      await ProxyService.check.with({ signal: controller.signal })(settingsFrom(draft));
      if (!controller.signal.aborted) toast.add({ title: "连接成功", type: "success" });
    } catch (err) {
      if (!controller.signal.aborted) {
        toast.add({ title: "连接失败", description: messageFrom(err, "连接检查失败。"), type: "error" });
      }
    } finally {
      if (activeRequest.current === controller) {
        activeRequest.current = null;
        setAction(null);
      }
    }
  }

  function cancelCheck() {
    const request = activeRequest.current;
    if (action !== "checking" || !(request instanceof AbortController)) return;
    request.abort();
    activeRequest.current = null;
    setAction(null);
  }

  return {
    settings, action, load, change, check, cancelCheck,
    save: () => { if (draft) return writeSettings(draft); },
    reset: () => writeSettings(null),
  };
}
