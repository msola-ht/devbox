import type { LucideIcon } from "lucide-react";

export interface NavigationLink {
  title: string;
  url: string;
}

export interface NavigationGroup {
  title: string;
  icon: LucideIcon;
  isActive?: boolean;
  items?: readonly NavigationLink[];
}

export interface Project {
  name: string;
  url: string;
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
  teams: readonly Team[];
  navigation: readonly NavigationGroup[];
  projects: readonly Project[];
}
