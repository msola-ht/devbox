import { createHashRouter } from "react-router-dom"

import { AppLayout } from "@/components/app-layout"
import { DashboardPage } from "@/pages/dashboard"
import { NotFoundPage } from "@/pages/not-found"
import { SettingsPage } from "@/pages/settings"
import { ToolsPage } from "@/pages/tools"

export const router = createHashRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      {
        index: true,
        element: <DashboardPage />,
        handle: { crumb: "概览" },
      },
      {
        path: "tools",
        element: <ToolsPage />,
        handle: { crumb: "工具" },
      },
      {
        path: "settings",
        element: <SettingsPage />,
        handle: { crumb: "设置" },
      },
      {
        path: "*",
        element: <NotFoundPage />,
        handle: { crumb: "页面未找到" },
      },
    ],
  },
])
