import {
  GalleryVerticalEndIcon,
  LayoutDashboardIcon,
  Settings2Icon,
  WrenchIcon,
} from "lucide-react";

import { appRoutes } from "@/app-routes";

import type { SidebarData } from "./model";

export const navigationData = {
  user: {
    name: "shadcn",
    email: "m@example.com",
  },
  team: {
    name: "Acme Inc",
    icon: GalleryVerticalEndIcon,
    plan: "Enterprise",
  },
  navigation: [
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
  ],
} satisfies SidebarData;
