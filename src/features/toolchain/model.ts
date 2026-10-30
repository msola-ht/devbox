import type {
  InstallationSummary,
  ToolSummary,
} from "@/features/toolchain/contracts"

export interface ToolStatus {
  label: string
  active: boolean
}

export function getToolStatus(tool: ToolSummary): ToolStatus {
  const healthyManagedInstallations = tool.managedInstallations.filter(
    (installation) => installation.healthy
  )
  const activeManagedInstallation = healthyManagedInstallations.find(
    (installation) => installation.installationId === tool.activeInstallationId
  )

  if (activeManagedInstallation) {
    return { label: "DevBox 已激活", active: true }
  }

  if (healthyManagedInstallations.length > 0) {
    return { label: "已安装，未激活", active: false }
  }

  if (tool.managedInstallations.length > 0) {
    return { label: "托管版本异常", active: false }
  }

  if (tool.systemInstallations.length > 0) {
    const healthySystemInstallations = tool.systemInstallations.filter(
      (installation) => installation.healthy
    )
    return {
      label:
        healthySystemInstallations.length > 0
          ? "检测到系统版本"
          : "系统版本异常",
      active: false,
    }
  }

  return { label: "未安装", active: false }
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
