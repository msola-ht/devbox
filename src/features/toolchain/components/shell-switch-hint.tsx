import { TerminalIcon } from "lucide-react"

import type { ToolSummary } from "@/features/toolchain/contracts"
import { Badge } from "@/components/ui/badge"

export function ShellSwitchHint({ shells }: { shells: ToolSummary[] }) {
  if (shells.length === 0) {
    return null
  }

  return (
    <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
      <TerminalIcon className="size-3.5" />
      <span>检测到可切换 Shell：</span>
      {shells.map((shell) => (
        <Badge key={shell.toolId} variant="outline">
          {shell.displayName} 可切换
        </Badge>
      ))}
      <span>切换能力待接入</span>
    </div>
  )
}
