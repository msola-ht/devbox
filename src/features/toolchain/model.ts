import type {
  InstallationSummary,
  ToolSummary,
} from "@/features/toolchain/contracts"

const SHELL_TOOL_IDS = new Set(["bash", "zsh"])

export interface ToolStatus {
  label: string | null
  active: boolean
}

function shellToolId(defaultShell: string | null) {
  if (!defaultShell) {
    return null
  }

  return (
    defaultShell
      .split(/[\\/]/)
      .filter(Boolean)
      .at(-1)
      ?.replace(/\.exe$/i, "") ?? null
  )
}

export function isDefaultShellTool(
  tool: ToolSummary,
  defaultShell: string | null
) {
  return shellToolId(defaultShell) === tool.toolId
}

export function filterToolsForDefaultShell(
  tools: ToolSummary[],
  defaultShell: string | null
) {
  if (!defaultShell) {
    return tools.filter((tool) => !SHELL_TOOL_IDS.has(tool.toolId))
  }

  const currentShellId = shellToolId(defaultShell)
  return tools.filter(
    (tool) => !SHELL_TOOL_IDS.has(tool.toolId) || tool.toolId === currentShellId
  )
}

export function getToolStatus(tool: ToolSummary): ToolStatus {
  const current = getCurrentInstallation(tool)
  if (!current) {
    return {
      label: tool.managedInstallations.length > 0 ? "无生效版本" : "未安装",
      active: false,
    }
  }

  if (!current.healthy) {
    return {
      label: current.source === "system" ? "系统版本异常" : "托管版本异常",
      active: false,
    }
  }

  if (current.source === "managed") {
    return { label: "DevBox 已激活", active: true }
  }

  return { label: null, active: false }
}

export function getCurrentInstallation(tool: ToolSummary) {
  return [...tool.systemInstallations, ...tool.managedInstallations].find(
    (installation) => installation.installationId === tool.currentInstallationId
  )
}

export function isActiveManagedInstallation(
  tool: ToolSummary,
  installation: InstallationSummary
) {
  return (
    installation.source === "managed" &&
    installation.healthy &&
    installation.installationId === tool.activeInstallationId
  )
}

export function installationKey(installation: InstallationSummary) {
  return `${installation.source}:${installation.installationId}:${installation.path}`
}
