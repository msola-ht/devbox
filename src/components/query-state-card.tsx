import { TriangleAlertIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

export function QueryLoadingCard({
  title,
  description = "正在读取本机状态",
}: {
  title: string
  description?: string
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <Skeleton className="h-4 w-2/3" />
        <Skeleton className="h-4 w-1/2" />
        <Skeleton className="h-10 w-full" />
      </CardContent>
    </Card>
  )
}

export function QueryErrorCard({
  title,
  code,
  message,
  onRetry,
}: {
  title: string
  code: string
  message: string
  onRetry: () => void
}) {
  return (
    <Card className="border-destructive/30">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TriangleAlertIcon className="size-4 text-destructive" />
          {title}
        </CardTitle>
        <CardDescription>
          {code} · {message}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Button variant="outline" onClick={onRetry}>
          重新读取
        </Button>
      </CardContent>
    </Card>
  )
}

export function QueryErrorNotice({
  code,
  message,
  suffix,
}: {
  code: string
  message: string
  suffix?: string
}) {
  return (
    <div className="rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2 text-xs text-destructive">
      {code} · {message}
      {suffix ? `，${suffix}` : null}
    </div>
  )
}
