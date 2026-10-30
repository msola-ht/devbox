import { PackageOpenIcon } from "lucide-react"

import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty"

export function ToolchainEmptyState() {
  return (
    <Empty className="border">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <PackageOpenIcon />
        </EmptyMedia>
        <EmptyTitle>没有已注册工具</EmptyTitle>
        <EmptyDescription>
          当前插件注册表为空，检测流程本身已完成。
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  )
}
