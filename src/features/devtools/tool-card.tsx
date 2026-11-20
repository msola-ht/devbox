import type { Cancellable } from "mygo-runtime";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from "@/components/ui/item";

import type { ToolDetection, ToolStatus } from "@/mygo";

import { useToolDetection } from "./use-tool-detection";

const statusLabels: Record<ToolStatus, string> = {
  available: "可用",
  not_found: "当前环境未找到",
  unsupported: "当前平台不适用",
  failed: "检测失败",
};

interface ToolCardProps {
  name: string;
  description: string;
  detect: Cancellable<[], ToolDetection>;
}

export function ToolCard({ name, description, detect }: ToolCardProps) {
  const { state, refresh } = useToolDetection(detect);
  const result = typeof state === "string" ? undefined : state;
  const status = state === "request_failed" ? "failed" : result?.status;
  const label = status ? statusLabels[status] : "检测中…";
  const title =
    result?.status === "available" ? `${label} · ${result.version}` : label;
  const message =
    state === "request_failed" ? "暂时无法读取检测结果，请重试。" : result?.message;
  const action = status === "failed" ? "重试" : "刷新";

  return (
    <Card>
      <CardHeader>
        <CardTitle>{name}</CardTitle>
        <CardDescription>{description}</CardDescription>
        {status !== "unsupported" && (
          <CardAction>
            <Button
              type="button"
              variant="outline"
              size="sm"
              aria-label={`${action} ${name} 检测`}
              disabled={state === "detecting"}
              onClick={() => void refresh()}
            >
              {action}
            </Button>
          </CardAction>
        )}
      </CardHeader>
      <CardContent>
        {status === "failed" ? (
          <Alert variant="destructive">
            <AlertTitle>{title}</AlertTitle>
            <AlertDescription className="wrap-anywhere">
              <div>{message}</div>
              {result?.path && <code>{result.path}</code>}
            </AlertDescription>
          </Alert>
        ) : (
          <Item className="p-0" role="status">
            <ItemContent className="min-w-0">
              <ItemTitle>{title}</ItemTitle>
              {message && <ItemDescription>{message}</ItemDescription>}
              {result?.path && (
                <code className="wrap-anywhere">{result.path}</code>
              )}
            </ItemContent>
          </Item>
        )}
      </CardContent>
    </Card>
  );
}
