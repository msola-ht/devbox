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
import { TeamSwitcher } from "@/features/navigation/team-switcher";

import { navigationData } from "./data";

export function AppSidebar() {
  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <TeamSwitcher teams={navigationData.teams} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navigationData.navigation} />
        <NavProjects projects={navigationData.projects} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={navigationData.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
