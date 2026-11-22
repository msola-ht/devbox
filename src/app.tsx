import { useEffect, type ComponentType } from "react";
import { Navigate, useLocation } from "react-router";

import { appRoutes, isAppPath, saveLastRoute, type AppPath } from "@/app-routes";
import { DevToolsPage } from "@/features/devtools/dev-tools-page";
import { AppSidebar } from "@/features/navigation/app-sidebar";
import { PageBreadcrumb } from "@/features/navigation/page-breadcrumb";
import { SettingsPage } from "@/features/settings/settings-page";
import { AppShell } from "@/features/shell/app-shell";
import { WorkspacePage } from "@/features/workspace/workspace-page";

const pageComponents: Record<AppPath, ComponentType> = {
  "/": WorkspacePage,
  "/tools": DevToolsPage,
  "/settings": SettingsPage,
};

export function App() {
  const { pathname } = useLocation();

  useEffect(() => {
    if (isAppPath(pathname)) saveLastRoute(pathname);
  }, [pathname]);

  if (!isAppPath(pathname)) return <Navigate to="/" replace />;

  const route = appRoutes[pathname];
  const Page = pageComponents[pathname];

  return (
    <AppShell
      sidebar={<AppSidebar />}
      breadcrumb={<PageBreadcrumb items={[...route.parents, route.title]} />}
    >
      <section className="grid w-full gap-4">
        <Page />
      </section>
    </AppShell>
  );
}
