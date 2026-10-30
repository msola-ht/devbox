import { createHashRouter } from "react-router-dom"

import { AppLayout } from "@/components/app-layout"
import { DashboardPage } from "@/pages/dashboard"
import { InDevelopmentPage } from "@/pages/in-development"
import { NotFoundPage } from "@/pages/not-found"
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
        element: <InDevelopmentPage title="设置" />,
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
