import { Link } from "react-router-dom"

import { Page } from "@/components/page"
import { Button } from "@/components/ui/button"

export function NotFoundPage() {
  return (
    <Page className="items-center justify-center text-center">
      <div className="space-y-2">
        <p className="text-sm font-medium text-muted-foreground">404</p>
        <h1 className="font-heading text-3xl font-semibold tracking-tight">
          页面不存在
        </h1>
        <p className="text-sm text-muted-foreground">
          请检查地址，或返回全局工作台。
        </p>
      </div>
      <Button render={<Link to="/" />}>返回控制台</Button>
    </Page>
  )
}
