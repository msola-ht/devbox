import { BoxesIcon, LayoutDashboardIcon, Settings2Icon } from "lucide-react"
import { NavLink, useLocation } from "react-router-dom"

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

const overviewItems = [
  { title: "概览", to: "/", icon: LayoutDashboardIcon },
  { title: "工具", to: "/tools", icon: BoxesIcon },
  { title: "设置", to: "/settings", icon: Settings2Icon },
]

export function AppNavigation() {
  const { pathname } = useLocation()

  return (
    <>
      <SidebarGroup>
        <SidebarGroupLabel>工作区</SidebarGroupLabel>
        <SidebarMenu className="gap-[2.5px]">
          {overviewItems.map((item) => (
            <SidebarMenuItem key={item.to}>
              <SidebarMenuButton
                tooltip={item.title}
                isActive={pathname === item.to}
                render={<NavLink to={item.to} end />}
              >
                <item.icon />
                <span>{item.title}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroup>
    </>
  )
}
