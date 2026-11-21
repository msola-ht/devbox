import { RefreshCwIcon } from "lucide-react";
import type { Cancellable } from "mygo-runtime";
import { Suspense, useEffect, type ComponentType } from "react";
import { Link } from "react-router";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Item, ItemContent, ItemTitle } from "@/components/ui/item";
import { Spinner } from "@/components/ui/spinner";
import { toast } from "@/components/ui/toast";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

import { useRefreshable } from "@/hooks/use-refreshable";
import { ProjectDialog, type ToolDetection } from "@/mygo";

import { ToolResult, ToolResultsDialog } from "./tool-results";

interface ToolCardProps {
  name: string;
  description: string;
  detect: Cancellable<[boolean], ToolDetection[]>;
  project?: string;
  InstallAction?: ComponentType<{
    detection: ToolDetection | undefined;
    ready: boolean;
    project: string;
    reload: () => Promise<void>;
  }>;
}

export function ToolCard({ name, description, detect, project = "", InstallAction }: ToolCardProps) {
  const { data, pending, failed: requestFailed, refresh, reload } = useRefreshable(detect);
  const results = data ?? [];
  const detecting = pending !== null;
  const allFailed =
    results.length > 0 && results.every(({ status }) => status === "failed");
  const action = requestFailed || allFailed ? "重试" : "刷新";
  const refreshLabel = detecting ? `正在检测 ${name}` : `${action} ${name} 检测`;
  const summary =
    results.find(({ status }) => status === "available") ?? results[0];
  const version = summary?.status === "available" ? summary.version : "";
  const shortVersion = version.replace(/\s+\([^()]*\)$/, "");

  useEffect(() => {
    if (!project) return;
    function recheckProject() {
      void reload();
    }
    window.addEventListener("focus", recheckProject);
    return () => window.removeEventListener("focus", recheckProject);
  }, [project, reload]);

  async function openProjectDirectory() {
    try {
      await ProjectDialog.openDirectory(project);
    } catch (error) {
      toast.add({
        title: "无法打开工作目录",
        description: error instanceof Error ? error.message : undefined,
        type: "error",
      });
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="min-w-0" aria-live="polite">
          <Tooltip>
            <TooltipTrigger render={<span className="block truncate" tabIndex={0} />}>
              {shortVersion ? `${name} ${shortVersion}` : name}
            </TooltipTrigger>
            <TooltipContent className="wrap-anywhere">
              <div className="flex flex-col">
                <span>{description}</span>
                {version && <span>{version}</span>}
              </div>
            </TooltipContent>
          </Tooltip>
        </CardTitle>
        {summary?.status !== "unsupported" && (
          <CardAction className="flex items-center gap-1">
            {results.length > 1 && (
              <ToolResultsDialog name={name} results={results} />
            )}
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    type="button"
                    variant="outline"
                    size="icon-sm"
                    aria-label={refreshLabel}
                    disabled={detecting}
                    onClick={() => void refresh()}
                  />
                }
              >
                {detecting ? <Spinner aria-hidden /> : <RefreshCwIcon aria-hidden />}
              </TooltipTrigger>
              <TooltipContent>{refreshLabel}</TooltipContent>
            </Tooltip>
          </CardAction>
        )}
      </CardHeader>
      <CardContent
        className="flex flex-col gap-2"
        aria-live="polite"
        aria-busy={detecting}
      >
        {requestFailed && (
          <Alert variant="destructive">
            <AlertDescription>
              {summary
                ? "刷新失败，当前显示上次检测结果。"
                : "暂时无法读取检测结果，请重试。"}
            </AlertDescription>
          </Alert>
        )}
        {summary && <ToolResult result={summary} compact project={!!project} />}
        {project && summary?.status === "project_unavailable" && (
          <Button className="self-start" size="sm" variant="outline" nativeButton={false}
            render={<Link to="/" />}>
            前往概览
          </Button>
        )}
        {project && summary?.status === "failed" && (
          <Button className="self-start" type="button" size="sm" variant="outline"
            onClick={() => void openProjectDirectory()}>
            打开工作目录
          </Button>
        )}
        {detecting && !summary && (
          <Item className="p-0">
            <ItemContent>
              <ItemTitle>检测中…</ItemTitle>
            </ItemContent>
          </Item>
        )}
      </CardContent>
      {InstallAction && (
        <Suspense fallback={null}>
          <InstallAction
            detection={summary}
            ready={!detecting && !requestFailed}
            project={project}
            reload={reload}
          />
        </Suspense>
      )}
    </Card>
  );
}
