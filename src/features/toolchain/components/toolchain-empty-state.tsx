import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export function ToolchainEmptyState() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>没有已注册工具</CardTitle>
        <CardDescription>
          当前插件注册表为空，检测流程本身已完成。
        </CardDescription>
      </CardHeader>
    </Card>
  )
}
