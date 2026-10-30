import { Page } from "@/components/page"

export function InDevelopmentPage({ title }: { title: string }) {
  return (
    <Page className="items-center justify-center text-center">
      <div className="flex flex-col gap-2">
        <h1 className="text-xl font-semibold">{title}</h1>
        <p className="text-sm text-muted-foreground">开发中</p>
      </div>
    </Page>
  )
}
