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

import { useHomebrew } from "./use-homebrew";

export function DevToolsPage() {
  const { result, refresh } = useHomebrew();

  let title = "检测中…";
  let description = "";
  if (result) {
    switch (result.status) {
      case "available":
        title = `可用 · ${result.version}`;
        break;
      case "not_found":
        title = "当前环境未找到";
        description = "已检查应用的 PATH 和 macOS 默认安装路径。";
        break;
      case "unsupported":
        title = "当前平台不适用";
        description = "此项检测适用于 macOS。";
        break;
      case "failed":
        title = "检测失败";
        description = result.message;
        break;
    }
  }
  const failed = result?.status === "failed";

  return (
    <Card>
      <CardHeader>
        <CardTitle>Homebrew</CardTitle>
        <CardDescription>macOS 包管理器</CardDescription>
        {result?.status !== "unsupported" && (
          <CardAction>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={!result}
              onClick={() => void refresh()}
            >
              {failed ? "重试" : "刷新"}
            </Button>
          </CardAction>
        )}
      </CardHeader>
      <CardContent>
        {failed ? (
          <Alert variant="destructive">
            <AlertTitle>{title}</AlertTitle>
            <AlertDescription className="wrap-anywhere">
              <div>{description}</div>
              {result?.path && <code>{result.path}</code>}
            </AlertDescription>
          </Alert>
        ) : (
          <Item className="p-0" role="status">
            <ItemContent className="min-w-0">
              <ItemTitle>{title}</ItemTitle>
              {description && <ItemDescription>{description}</ItemDescription>}
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
