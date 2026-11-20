import { getStoredValue, setStoredValue } from "@/lib/storage";

export const appRoutes = {
  "/": { title: "概览", parents: [] },
  "/tools": { title: "开发工具", parents: [] },
  "/settings": { title: "设置", parents: [] },
} as const;

export type AppPath = keyof typeof appRoutes;

const lastRouteKey = "devbox:last-route";

export function isAppPath(path: string): path is AppPath {
  return Object.hasOwn(appRoutes, path);
}

export function restoreLastRoute() {
  if (window.location.hash) return;

  const savedPath = getStoredValue(lastRouteKey);
  if (savedPath && isAppPath(savedPath)) {
    window.history.replaceState(window.history.state, "", `#${savedPath}`);
  }
}

export function saveLastRoute(path: AppPath) {
  setStoredValue(lastRouteKey, path);
}
