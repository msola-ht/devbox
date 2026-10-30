export type InstallationSource = "system" | "managed"
export type ToolKind = "tool" | "shell"

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
  kind: ToolKind
  isDefaultShell: boolean
  effectiveInstallation: InstallationSummary | null
  managedInstallations: InstallationSummary[]
  activeInstallationId: string | null
  activeInstallationErrorCode: string | null
}

export interface ToolListSnapshot {
  bootId: string
  revision: string
  environmentRevision: string
  generatedAt: string
  tools: ToolSummary[]
}
