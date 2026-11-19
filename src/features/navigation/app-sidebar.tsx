import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar";
import { NavMain } from "@/features/navigation/nav-main";
import { NavProjects } from "@/features/navigation/nav-projects";
import { NavUser } from "@/features/navigation/nav-user";
import { TeamHeader } from "@/features/navigation/team-header";

import { navigationData } from "./data";

export function AppSidebar() {
  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <TeamHeader team={navigationData.team} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navigationData.navigation} />
        <NavProjects />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={navigationData.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
