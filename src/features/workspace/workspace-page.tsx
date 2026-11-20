import { EnvironmentCard } from "@/features/runtime-env/environment-card";
import { SystemInfoCard } from "@/features/system/system-info-card";

export function WorkspacePage() {
  return (
    <div className="@container">
      <div className="grid gap-4 @3xl:grid-cols-2">
        <SystemInfoCard />
        <EnvironmentCard />
      </div>
    </div>
  );
}
