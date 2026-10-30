import type {
  InstallationSummary,
  ToolSummary,
} from "@/features/toolchain/contracts"

export interface ToolStatus {
  label: string | null
  active: boolean
}

export function selectVisibleTools(tools: ToolSummary[]) {
  return tools.filter((tool) => tool.kind !== "shell" || tool.isDefaultShell)
}

export function selectSwitchableShells(tools: ToolSummary[]) {
  return tools.filter(
    (tool) =>
      tool.kind === "shell" &&
      !tool.isDefaultShell &&
      tool.shellInstallations.some((installation) => installation.healthy)
  )
}

export function getToolStatus(tool: ToolSummary): ToolStatus {
  if (tool.activeInstallationErrorCode) {
    return { label: "激活状态异常", active: false }
  }

  const current = tool.effectiveInstallation
  if (!current) {
    if (tool.activeInstallationId) {
      return { label: "已激活但不在当前 PATH", active: false }
    }

    if (tool.kind === "shell" && tool.shellInstallations.length > 0) {
      return { label: "默认路径不可用", active: false }
    }

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

  if (current.installationId === tool.activeInstallationId) {
    return { label: "DevBox 已激活", active: true }
  }

  if (tool.activeInstallationId) {
    return { label: "DevBox 已选择其他版本", active: false }
  }

  return { label: null, active: false }
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

export function isDefaultShellInstallation(
  tool: ToolSummary,
  installation: InstallationSummary
) {
  return (
    tool.kind === "shell" &&
    installation.installationId === tool.effectiveInstallation?.installationId
  )
}

export function installationKey(installation: InstallationSummary) {
  return `${installation.source}:${installation.installationId}:${installation.path}`
}
