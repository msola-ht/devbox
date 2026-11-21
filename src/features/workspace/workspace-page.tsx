import { isMyGo } from "mygo-runtime";
import { lazy, Suspense } from "react";

import { Button } from "@/components/ui/button";
import { EnvironmentCard } from "@/features/runtime-env/environment-card";
import { SystemInfoCard } from "@/features/system/system-info-card";
import { useRefreshable } from "@/hooks/use-refreshable";
import { EnvironmentService, SystemService } from "@/mygo";

const ProjectDirectoryBar = lazy(() => import("./project-directory-bar"));

export function WorkspacePage() {
  const system = useRefreshable(SystemService.get);
  const environment = useRefreshable(EnvironmentService.get);
  const refreshing = system.pending === "refreshing" || environment.pending === "refreshing";
  const busy = system.pending !== null || environment.pending !== null;

  function refresh() {
    void system.refresh();
    void environment.refresh();
  }

  return (
    <div className="@container flex flex-col gap-4">
      {isMyGo() && (
        <Suspense fallback={null}>
          <ProjectDirectoryBar />
        </Suspense>
      )}
      <div className="grid gap-4 @3xl:grid-cols-2">
        <SystemInfoCard
          snapshot={system.data}
          busy={system.pending !== null}
          failed={system.failed}
          action={
            <Button type="button" variant="outline" size="sm" disabled={busy} onClick={refresh}>
              {refreshing ? "刷新中…" : busy ? "读取中…" : "刷新概览"}
            </Button>
          }
        />
        <EnvironmentCard
          info={environment.data}
          busy={environment.pending !== null}
          failed={environment.failed}
        />
      </div>
    </div>
  );
}
