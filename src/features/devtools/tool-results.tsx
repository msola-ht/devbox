import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from "@/components/ui/item";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import type { ToolDetection, ToolStatus } from "@/mygo";

const statusLabels: Record<Exclude<ToolStatus, "available">, string> = {
  not_found: "当前环境未找到",
  unsupported: "当前平台不适用",
  failed: "检测失败",
};

export function ToolResultsDialog({
  name,
  results,
}: {
  name: string;
  results: ToolDetection[];
}) {
  const label = `查看 ${name} 的 ${results.length} 条检测结果`;

  return (
    <Dialog>
      <Tooltip>
        <DialogTrigger
          aria-label={label}
          render={
            <TooltipTrigger render={<Button variant="secondary" size="icon-sm" />} />
          }
        >
          {results.length}
        </DialogTrigger>
        <TooltipContent>{label}</TooltipContent>
      </Tooltip>
      <DialogContent className="max-h-[calc(100dvh-2rem)] grid-rows-[auto_minmax(0,1fr)] sm:max-w-lg">
        <DialogHeader className="min-h-7 justify-center pr-8">
          <DialogTitle>{name} 安装详情</DialogTitle>
        </DialogHeader>
        <ItemGroup
          className="min-h-0 overflow-y-auto"
          aria-label={`${name} 检测结果`}
          tabIndex={0}
        >
          {results.map((result) => (
            <ToolResult
              key={JSON.stringify([result.path, result.status, result.version])}
              result={result}
            />
          ))}
        </ItemGroup>
      </DialogContent>
    </Dialog>
  );
}

export function ToolResult({
  result,
  compact = false,
}: {
  result: ToolDetection;
  compact?: boolean;
}) {
  const title =
    result.status === "available"
      ? compact ? "" : result.version
      : statusLabels[result.status];
  const role = compact ? undefined : "listitem";
  const path = result.path && (
    compact ? (
      <Tooltip>
        <TooltipTrigger render={<code className="block truncate" tabIndex={0} />}>
          {result.path}
        </TooltipTrigger>
        <TooltipContent>
          <code className="min-w-0 wrap-anywhere">{result.path}</code>
        </TooltipContent>
      </Tooltip>
    ) : (
      <code className="wrap-anywhere">{result.path}</code>
    )
  );

  if (result.status === "failed") {
    return (
      <div role={role}>
        <Alert variant="destructive">
          <AlertTitle>{title}</AlertTitle>
          <AlertDescription className="min-w-0 wrap-anywhere">
            <div>{result.message}</div>
            {path}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <Item className="p-0" role={role}>
      <ItemContent className="min-w-0">
        {title && (
          <ItemTitle className="line-clamp-none max-w-full wrap-anywhere">
            {title}
          </ItemTitle>
        )}
        {result.message && <ItemDescription>{result.message}</ItemDescription>}
        {path}
      </ItemContent>
    </Item>
  );
}
