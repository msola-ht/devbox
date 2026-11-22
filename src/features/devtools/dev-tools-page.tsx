import { isMyGo, runtime } from "mygo-runtime";
import { useCallback, useEffect, useState } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DevToolsService, type InstallChannel, type InstallTool } from "@/mygo";

import { ToolCard } from "./tool-card";
import { getDevelopmentToolGroups, type ToolGroupDefinition } from "./tool-catalog";

// Reserve three 1rem gaps when sizing four cards across a row.
const toolRow =
  "flex flex-wrap gap-4 *:flex-1 *:min-w-[min(100%,max(14rem,calc((100%_-_3rem)/4)))]";

function ToolGroup({ id, title, tools, channels }: ToolGroupDefinition & {
  channels: ReadonlyMap<InstallTool, InstallChannel>;
}) {
  const [runtimeRevision, setRuntimeRevision] = useState({ node: 0, npm: 0 });
  const [nodeAvailable, setNodeAvailable] = useState(false);
  const onInstallSettled = useCallback((tool: InstallTool) => {
    if (tool === "node" || tool === "npm") {
      const companion = tool === "node" ? "npm" : "node";
      setRuntimeRevision((current) => ({ ...current, [companion]: current[companion] + 1 }));
    }
  }, []);

  return (
    <section aria-labelledby={id} className="flex min-w-0 flex-col gap-3">
      <h2 id={id} className="text-sm font-medium">{title}</h2>
      <div className={toolRow}>
        {tools.map((tool) => (
          <ToolCard key={tool.id} {...tool} channel={channels.get(tool.id)} onInstallSettled={onInstallSettled}
            hideInstallButton={tool.id === "npm" && !nodeAvailable}
            onAvailabilityChange={tool.id === "node" ? setNodeAvailable : undefined}
            recheckRevision={tool.id === "node" || tool.id === "npm" ? runtimeRevision[tool.id] : 0} />
        ))}
      </div>
    </section>
  );
}

export function DevToolsPage() {
  const platform = isMyGo() ? runtime().platform : undefined;
  const groups = getDevelopmentToolGroups(platform);
  const [channels, setChannels] = useState<ReadonlyMap<InstallTool, InstallChannel>>(new Map());
  const [installOptionsFailed, setInstallOptionsFailed] = useState(false);
  const [installOptionsRevision, setInstallOptionsRevision] = useState(0);

  useEffect(() => {
    if (!platform) return;
    const controller = new AbortController();
    void DevToolsService.getToolInstallOptions.with({ signal: controller.signal })().then((options) => {
      if (!controller.signal.aborted) {
        setChannels(new Map(options.map(({ tool, channel }) => [tool, channel])));
        setInstallOptionsFailed(false);
      }
    }).catch(() => {
      if (!controller.signal.aborted) setInstallOptionsFailed(true);
    });
    return () => controller.abort();
  }, [platform, installOptionsRevision]);
  const platformGroups = groups.filter(({ id }) =>
    id === "tools-system-packages" || id === "tools-shells",
  );
  const showTogether = platformGroups.length === 2;
  const otherGroups = showTogether ? groups.filter((group) => !platformGroups.includes(group)) : groups;

  return (
    <div className="@container flex flex-col gap-6">
      {installOptionsFailed && (
        <Alert variant="destructive">
          <AlertDescription className="flex flex-wrap items-center gap-2">
            无法读取安装选项。
            <Button type="button" size="sm" variant="outline"
              onClick={() => setInstallOptionsRevision((revision) => revision + 1)}>
              重试
            </Button>
          </AlertDescription>
        </Alert>
      )}
      {showTogether && (
        <div className="grid items-start gap-6 @min-[30rem]:grid-cols-2">
          {platformGroups.map((group) => (
            <ToolGroup key={group.id} {...group} channels={channels} />
          ))}
        </div>
      )}
      {otherGroups.map((group) => (
        <ToolGroup key={group.id} {...group} channels={channels} />
      ))}
    </div>
  );
}
