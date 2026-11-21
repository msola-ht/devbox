import { RefreshCwIcon } from "lucide-react";
import type { Cancellable } from "mygo-runtime";

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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

import { useRefreshable } from "@/hooks/use-refreshable";
import type { ToolDetection } from "@/mygo";

import { ToolResult, ToolResultsDialog } from "./tool-results";

interface ToolCardProps {
  name: string;
  description: string;
  detect: Cancellable<[boolean], ToolDetection[]>;
}

export function ToolCard({ name, description, detect }: ToolCardProps) {
  const { data, pending, failed: requestFailed, refresh } = useRefreshable(detect);
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
        {summary && <ToolResult result={summary} compact />}
        {detecting && !summary && (
          <Item className="p-0">
            <ItemContent>
              <ItemTitle>检测中…</ItemTitle>
            </ItemContent>
          </Item>
        )}
      </CardContent>
    </Card>
  );
}
