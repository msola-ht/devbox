import { RefreshCwIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"

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
      {loading ? (
        <Spinner data-icon="inline-start" />
      ) : (
        <RefreshCwIcon data-icon="inline-start" />
      )}
      {label}
    </Button>
  )
}
