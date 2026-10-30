import type { ReactNode } from "react"
import { cn } from "cn"

export function Page({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <div className={cn("flex w-full flex-1 flex-col gap-6", className)}>
      {children}
    </div>
  )
}
