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
import type { ToolDetection, ToolSource, ToolStatus } from "@/mygo";

const statusLabels: Record<Exclude<ToolStatus, "available">, string> = {
  not_found: "当前环境未找到",
  unsupported: "当前平台不适用",
  failed: "检测失败",
  incompatible: "架构或系统不匹配",
  project_unavailable: "工作目录不可用",
};

const sourceLabels: Record<Exclude<ToolSource, "">, string> = {
  unknown: "未知",
  system: "系统自带",
  apple_developer_tools: "Apple 开发工具",
  homebrew: "Homebrew",
  fnm: "fnm",
  nvm: "nvm",
  nvm_windows: "nvm-windows",
  uv: "uv",
  corepack: "Corepack",
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
      <DialogContent className="max-h-[calc(100dvh-2rem)] grid-cols-1 grid-rows-[auto_minmax(0,1fr)] sm:max-w-lg">
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
  if (compact && result.status === "not_found") {
    if (!result.path) return <span>{statusLabels.not_found}</span>;
    return (
      <Tooltip>
        <TooltipTrigger render={<span tabIndex={0} />}>
          {statusLabels.not_found}
        </TooltipTrigger>
        <TooltipContent>
          <div className="flex min-w-0 flex-col wrap-anywhere">
            {result.message && <span>{result.message}</span>}
            <code>{result.path}</code>
          </div>
        </TooltipContent>
      </Tooltip>
    );
  }

  let title =
    result.status === "available"
      ? result.version
      : statusLabels[result.status];
  if (result.status === "not_found" && result.path && result.message) {
    title = result.message;
  }
  const role = compact ? undefined : "listitem";
  const path = result.path && (
    <Tooltip>
      <TooltipTrigger render={<code className="block min-w-0 truncate" tabIndex={0} />}>
        {result.path}
      </TooltipTrigger>
      <TooltipContent>
        <code className="min-w-0 wrap-anywhere" aria-hidden="true">
          {result.path}
        </code>
      </TooltipContent>
    </Tooltip>
  );

  if (compact && result.status === "available") {
    return (
      <div className="flex min-w-0 items-baseline gap-x-1.5">
        {result.source && (
          <span className="shrink-0 text-muted-foreground">
            <span className="sr-only">安装来源 </span>
            {sourceLabels[result.source]}
            {path && <span aria-hidden> ·</span>}
          </span>
        )}
        {path}
      </div>
    );
  }

  if (result.status === "failed" || result.status === "incompatible" || result.status === "project_unavailable") {
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
        {result.status !== "not_found" && result.message && (
          <ItemDescription>{result.message}</ItemDescription>
        )}
        {path}
        {result.source && (
          <ItemDescription>安装来源 · {sourceLabels[result.source]}</ItemDescription>
        )}
      </ItemContent>
    </Item>
  );
}
