import type { LucideIcon } from "lucide-react";

import type { AppPath } from "@/app-routes";

export interface NavigationItem {
  title: string;
  url: AppPath;
  icon: LucideIcon;
}

export interface Team {
  name: string;
  icon: LucideIcon;
  plan: string;
}

export interface User {
  name: string;
  email: string;
  avatar?: string;
}

export interface SidebarData {
  user: User;
  team: Team;
  navigation: readonly NavigationItem[];
}
