import { DevToolsService } from "@/mygo";

import { ToolCard } from "./tool-card";

export function DevToolsPage() {
  return (
    <div className="grid grid-cols-[repeat(auto-fit,minmax(min(100%,14rem),1fr))] gap-4">
      <ToolCard
        name="Homebrew"
        description="macOS 包管理器"
        detect={DevToolsService.detectHomebrew}
      />
      <ToolCard
        name="Git"
        description="版本控制工具"
        detect={DevToolsService.detectGit}
      />
      <ToolCard
        name="uv"
        description="Python 包与项目管理工具"
        detect={DevToolsService.detectUV}
      />
      <ToolCard
        name="Python"
        description="Python 3 解释器"
        detect={DevToolsService.detectPython}
      />
    </div>
  );
}
