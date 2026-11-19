import { Link, useLocation } from "react-router";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";

import type { NavigationItem } from "./model";

export function NavMain({ items }: { items: readonly NavigationItem[] }) {
  const { pathname } = useLocation();
  const { setOpenMobile } = useSidebar();

  function handleNavigate() {
    setOpenMobile(false);
  }

  return (
    <SidebarGroup>
      <SidebarGroupLabel>控制台</SidebarGroupLabel>
      <SidebarMenu className="gap-px">
        {items.map((item) => {
          const Icon = item.icon;
          const isActive = item.url === pathname;

          return (
            <SidebarMenuItem key={item.url}>
              <SidebarMenuButton
                tooltip={item.title}
                isActive={isActive}
                aria-current={isActive ? "page" : undefined}
                render={<Link to={item.url} onClick={handleNavigate} />}
              >
                <Icon />
                <span>{item.title}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          );
        })}
      </SidebarMenu>
    </SidebarGroup>
  );
}
