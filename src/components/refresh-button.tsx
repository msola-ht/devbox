import { RefreshCwIcon } from "lucide-react"

import { Button } from "@/components/ui/button"

export function RefreshButton({
  loading,
  onClick,
  label = "刷新检测",
}: {
  loading: boolean
  onClick: () => void
  label?: string
}) {
  return (
    <Button variant="outline" onClick={onClick} disabled={loading}>
      <RefreshCwIcon
        data-icon="inline-start"
        className={loading ? "animate-spin" : undefined}
      />
      {label}
    </Button>
  )
}
