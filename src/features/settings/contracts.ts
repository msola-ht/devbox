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

export type SystemProxySource = "environment" | "operatingSystem"

export interface SystemProxyEndpoint {
  url: string
  source: SystemProxySource
  requiresAuthentication: boolean
}

export interface SystemProxyInspection {
  checkedAt: string
  http: SystemProxyEndpoint | null
  https: SystemProxyEndpoint | null
}

export interface SettingsSnapshot {
  bootId: string
  revision: string
  generatedAt: string
  proxy: ProxySettings
  systemProxy: SystemProxyInspection
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
