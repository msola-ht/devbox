import { AppSidebar } from "@/features/navigation/app-sidebar";
import { AppShell } from "@/features/shell/app-shell";
import { WorkspacePage } from "@/features/workspace/workspace-page";

export function App() {
  return (
    <AppShell sidebar={<AppSidebar />}>
      <WorkspacePage />
    </AppShell>
  );
}
