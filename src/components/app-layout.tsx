import { MoonIcon, SunIcon } from "lucide-react"
import { Link, Outlet, useMatches } from "react-router-dom"

import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Toaster } from "@/components/ui/toast"
import { TooltipProvider } from "@/components/ui/tooltip"
import { useTheme } from "@/components/theme-provider"

type RouteHandle = {
  crumb?: string
}

function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const isDark = theme === "dark"

  return (
    <Button
      variant="ghost"
      size="icon-sm"
      aria-label="切换主题"
      title="切换主题"
      onClick={() => setTheme(isDark ? "light" : "dark")}
    >
      {isDark ? <SunIcon /> : <MoonIcon />}
    </Button>
  )
}

export function AppLayout() {
  const matches = useMatches()
  const currentHandle = [...matches].reverse().find((match) => {
    const handle = match.handle as RouteHandle | undefined
    return Boolean(handle?.crumb)
  })?.handle as RouteHandle | undefined
  const title = currentHandle?.crumb ?? "工作台"

  return (
    <TooltipProvider>
      <SidebarProvider defaultOpen={false} className="h-svh overflow-hidden">
        <AppSidebar />
        <SidebarInset className="min-h-0 overflow-hidden">
          <header className="shrink-0 border-b">
            <div className="flex h-14 w-full items-center gap-2 px-4 md:px-6">
              <SidebarTrigger />
              <div className="flex h-4 items-center">
                <Separator orientation="vertical" />
              </div>
              <Breadcrumb>
                <BreadcrumbList>
                  <BreadcrumbItem className="hidden md:block">
                    <BreadcrumbLink render={<Link to="/" />}>
                      DevBox
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                  <BreadcrumbSeparator className="hidden md:block" />
                  <BreadcrumbItem>
                    <BreadcrumbPage>{title}</BreadcrumbPage>
                  </BreadcrumbItem>
                </BreadcrumbList>
              </Breadcrumb>
              <div className="ml-auto flex items-center gap-2">
                <ThemeToggle />
              </div>
            </div>
          </header>
          <ScrollArea className="min-h-0 flex-1">
            <div className="flex min-h-full w-full flex-col p-4 md:p-6">
              <Outlet />
            </div>
          </ScrollArea>
        </SidebarInset>
        <Toaster />
      </SidebarProvider>
    </TooltipProvider>
  )
}
