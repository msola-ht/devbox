import { TriangleAlertIcon } from "lucide-react"

import type { EnvironmentSnapshot } from "@/features/environment/contracts"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"

function shellEnvironmentMessage(
  status: EnvironmentSnapshot["shellEnvironmentStatus"]
) {
  switch (status) {
    case "defaultShellUnavailable":
      return "无法读取账户默认 Shell，当前不会把其他 PATH 当作默认 Shell 环境。"
    case "unsupportedShell":
      return "默认 Shell 暂不在 Bash、Zsh 支持范围内，无法读取对应的登录 PATH。"
    case "loginPathUnavailable":
      return "默认 Shell 登录环境读取失败或超时，当前不会回退到应用进程的 PATH。"
    case "ready":
      return null
  }
}

export function ShellEnvironmentNotice({
  environment,
}: {
  environment: Pick<
    EnvironmentSnapshot,
    "defaultShell" | "shellEnvironmentStatus"
  >
}) {
  const message = shellEnvironmentMessage(environment.shellEnvironmentStatus)
  if (!message) {
    return null
  }

  return (
    <Alert variant="destructive">
      <TriangleAlertIcon />
      <AlertTitle>默认 Shell 登录环境不可用</AlertTitle>
      <AlertDescription>
        {message}
        {environment.defaultShell
          ? ` 当前账户默认 Shell：${environment.defaultShell}。`
          : null}
      </AlertDescription>
    </Alert>
  )
}
