import {
  LayoutDashboardIcon,
  Settings2Icon,
  WrenchIcon,
} from "lucide-react";

import { appRoutes } from "@/app-routes";

import type { NavigationItem } from "./model";

export const navigationData = [
  {
    title: appRoutes["/"].title,
    icon: LayoutDashboardIcon,
    url: "/",
  },
  {
    title: appRoutes["/tools"].title,
    icon: WrenchIcon,
    url: "/tools",
  },
  {
    title: appRoutes["/settings"].title,
    icon: Settings2Icon,
    url: "/settings",
  },
] satisfies readonly NavigationItem[];
