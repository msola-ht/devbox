import type { LucideIcon } from "lucide-react";

import type { AppPath } from "@/app-routes";

export interface NavigationItem {
  title: string;
  url: AppPath;
  icon: LucideIcon;
}
