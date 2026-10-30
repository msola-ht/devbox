import type {
  ProxyProtocol,
  ProxySettings,
} from "@/features/settings/contracts"

export interface ProxyDraft {
  mode: "direct" | "manual"
  protocol: ProxyProtocol
  host: string
  port: string
  bypass: string
}

export function draftFromProxy(proxy: ProxySettings): ProxyDraft {
  if (proxy.mode === "direct") {
    return {
      mode: "direct",
      protocol: "http",
      host: "",
      port: "",
      bypass: "",
    }
  }

  return {
    mode: "manual",
    protocol: proxy.protocol,
    host: proxy.host,
    port: String(proxy.port),
    bypass: proxy.bypass.join(", "),
  }
}

export function draftToProxy(draft: ProxyDraft): ProxySettings | null {
  if (draft.mode === "direct") {
    return { mode: "direct" }
  }

  const port = Number(draft.port)
  if (!isValidPort(draft.port) || draft.host.trim() === "") {
    return null
  }

  return {
    mode: "manual",
    protocol: draft.protocol,
    host: draft.host.trim(),
    port,
    bypass: parseBypass(draft.bypass),
  }
}

export function isValidPort(value: string): boolean {
  if (!/^\d+$/.test(value)) {
    return false
  }
  const port = Number(value)
  return Number.isInteger(port) && port >= 1 && port <= 65535
}

export function parseBypass(value: string): string[] {
  return value
    .split(/[,\n]/)
    .map((entry) => entry.trim())
    .filter(Boolean)
}

export function proxySummary(proxy: ProxySettings): string {
  if (proxy.mode === "direct") {
    return "直连"
  }
  return `${proxy.protocol}://${proxy.host}:${proxy.port}`
}
