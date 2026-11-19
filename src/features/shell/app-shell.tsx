import { useState, type ReactNode } from "react";

import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ThemeToggle } from "@/features/theme/theme-toggle";
import { getStoredValue, setStoredValue } from "@/lib/storage";

const sidebarOpenKey = "devbox:sidebar-open";

interface AppShellProps {
  sidebar: ReactNode;
  breadcrumb: ReactNode;
  children: ReactNode;
}

export function AppShell({ sidebar, breadcrumb, children }: AppShellProps) {
  const [sidebarOpen, setSidebarOpen] = useState(
    () => getStoredValue(sidebarOpenKey) === "true",
  );

  function handleSidebarOpenChange(open: boolean) {
    setSidebarOpen(open);
    setStoredValue(sidebarOpenKey, String(open));
  }

  return (
    <TooltipProvider>
      <SidebarProvider open={sidebarOpen} onOpenChange={handleSidebarOpenChange}>
        {sidebar}
        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
            <div className="flex w-full items-center gap-2 px-4">
              <SidebarTrigger className="-ml-1" />
              <Separator
                orientation="vertical"
                className="mr-2 h-4 self-center"
              />
              {breadcrumb}
              <div className="ml-auto">
                <ThemeToggle />
              </div>
            </div>
          </header>
          <div className="flex flex-1 flex-col gap-4 p-4 pt-0">{children}</div>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  );
}
