export type ProxyProtocol = "http" | "socks5h"

export interface SystemProxySettings {
  mode: "system"
}

export interface DirectProxySettings {
  mode: "direct"
}

export interface ManualProxySettings {
  mode: "manual"
  protocol: ProxyProtocol
  host: string
  port: number
  bypass: string[]
}

export type ProxySettings =
  SystemProxySettings | DirectProxySettings | ManualProxySettings

export interface SettingsSnapshot {
  bootId: string
  revision: string
  generatedAt: string
  proxy: ProxySettings
}

export interface UpdateSettingsRequest {
  expectedRevision: string
  proxy: ProxySettings
}

export interface ProxyTestResult {
  checkedAt: string
  latencyMs: number
  httpStatus: number
}
