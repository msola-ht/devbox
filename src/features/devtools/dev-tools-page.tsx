import { ToolCard } from "./tool-card";
import { getDevelopmentToolGroups, type ToolGroupDefinition } from "./tool-catalog";

// Reserve three 1rem gaps when sizing four cards across a row.
const toolRow =
  "flex flex-wrap gap-4 *:flex-1 *:min-w-[min(100%,max(14rem,calc((100%_-_3rem)/4)))]";

function ToolGroup({ id, title, tools }: ToolGroupDefinition) {
  return (
    <section aria-labelledby={id} className="flex min-w-0 flex-col gap-3">
      <h2 id={id} className="text-sm font-medium">{title}</h2>
      <div className={toolRow}>
        {tools.map((tool) => <ToolCard key={tool.id} {...tool} />)}
      </div>
    </section>
  );
}

export function DevToolsPage() {
  const groups = getDevelopmentToolGroups();
  const platformGroups = groups.filter(({ id }) =>
    id === "tools-system-packages" || id === "tools-shells",
  );
  const showTogether = platformGroups.length === 2;
  const otherGroups = showTogether ? groups.filter((group) => !platformGroups.includes(group)) : groups;

  return (
    <div className="@container flex flex-col gap-6">
      {showTogether && (
        <div className="grid items-start gap-6 @min-[30rem]:grid-cols-2">
          {platformGroups.map((group) => <ToolGroup key={group.id} {...group} />)}
        </div>
      )}
      {otherGroups.map((group) => <ToolGroup key={group.id} {...group} />)}
    </div>
  );
}
