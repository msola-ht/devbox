export type InstallationSource = "system" | "managed"

export interface InstallationSummary {
  installationId: string
  version: string | null
  path: string
  source: InstallationSource
  sourceHint: string | null
  healthy: boolean
  errorCode: string | null
}

export interface ToolSummary {
  toolId: string
  displayName: string
  description: string
  pluginContractVersion: string
  systemInstallations: InstallationSummary[]
  managedInstallations: InstallationSummary[]
  activeInstallationId: string | null
}

export interface ToolListSnapshot {
  bootId: string
  revision: string
  generatedAt: string
  tools: ToolSummary[]
}
