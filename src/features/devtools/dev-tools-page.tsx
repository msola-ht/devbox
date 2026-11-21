import { isMyGo, runtime } from "mygo-runtime";
import { lazy, useMemo, useState } from "react";
import { Link } from "react-router";

import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { useProjectDirectory } from "@/lib/project-directory";
import { getStoredValue, setStoredValue } from "@/lib/storage";
import { DevToolsService } from "@/mygo";

import { ToolCard } from "./tool-card";
import { getDevelopmentToolGroups, type ToolGroupDefinition } from "./tool-catalog";

const UVInstallAction = lazy(() => import("./uv-install-action"));
const scopeKey = "devbox:tool-scope";

// Reserve three 1rem gaps when sizing four cards across a row.
const toolRow =
  "flex flex-wrap gap-4 *:flex-1 *:min-w-[min(100%,max(14rem,calc((100%_-_3rem)/4)))]";

function ToolGroup({ id, title, tools, supportsUVInstall }: ToolGroupDefinition & {
  supportsUVInstall: boolean;
}) {
  return (
    <section aria-labelledby={id} className="flex min-w-0 flex-col gap-3">
      <h2 id={id} className="text-sm font-medium">{title}</h2>
      <div className={toolRow}>
        {tools.map((tool) => (
          <ToolCard
            key={tool.id}
            {...tool}
            InstallAction={supportsUVInstall && tool.id === "uv" ? UVInstallAction : undefined}
          />
        ))}
      </div>
    </section>
  );
}

function ProjectTools({ project, supportsUVInstall }: { project: string; supportsUVInstall: boolean }) {
  const detect = useMemo(() => Object.assign(
    (refresh: boolean) => DevToolsService.detectProjectUV(project, refresh),
    { with: (options: Parameters<typeof DevToolsService.detectProjectUV.with>[0]) =>
      (refresh: boolean) => DevToolsService.detectProjectUV.with(options)(project, refresh) },
  ), [project]);

  return (
    <div className={toolRow}>
      <ToolCard
        name="uv"
        description="项目的 Python 包与项目管理工具"
        detect={detect}
        project={project}
        InstallAction={supportsUVInstall ? UVInstallAction : undefined}
      />
    </div>
  );
}

export function DevToolsPage() {
  const platform = isMyGo() ? runtime().platform : undefined;
  const [scope, setScope] = useState(() =>
    platform && getStoredValue(scopeKey) === "project" ? "project" : "system");
  const { directory: project, initializing } = useProjectDirectory();
  const groups = getDevelopmentToolGroups(platform);
  const supportsUVInstall = platform === "darwin";
  const platformGroups = groups.filter(({ id }) =>
    id === "tools-system-packages" || id === "tools-shells",
  );
  const showTogether = platformGroups.length === 2;
  const otherGroups = showTogether ? groups.filter((group) => !platformGroups.includes(group)) : groups;

  return (
    <div className="@container flex flex-col gap-6">
      {platform && (
        <div className="flex flex-wrap items-center gap-2">
          <ToggleGroup aria-label="工具作用域" value={[scope]} onValueChange={([value]) => {
            if (value) {
              setScope(value);
              setStoredValue(scopeKey, value);
            }
          }} size="sm" variant="outline">
            <ToggleGroupItem value="system">系统全局</ToggleGroupItem>
            <ToggleGroupItem value="project">项目隔离</ToggleGroupItem>
          </ToggleGroup>
        </div>
      )}
      {scope === "project" ? (
        <>
          {project ? (
            <>
              <code className="min-w-0 wrap-anywhere">{project}</code>
              <ProjectTools key={project} project={project} supportsUVInstall={supportsUVInstall} />
            </>
          ) : initializing ? (
            <p role="status" className="text-sm text-muted-foreground">正在准备默认工作目录…</p>
          ) : (
            <p className="text-sm text-muted-foreground">
              请先在<Link to="/" className="underline">概览</Link>选择并确认工作目录。
            </p>
          )}
        </>
      ) : (
        <>
          {showTogether && (
            <div className="grid items-start gap-6 @min-[30rem]:grid-cols-2">
              {platformGroups.map((group) => (
                <ToolGroup key={group.id} {...group} supportsUVInstall={supportsUVInstall} />
              ))}
            </div>
          )}
          {otherGroups.map((group) => (
            <ToolGroup key={group.id} {...group} supportsUVInstall={supportsUVInstall} />
          ))}
        </>
      )}
    </div>
  );
}
