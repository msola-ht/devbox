import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
} from "@/components/ui/item";
import { Skeleton } from "@/components/ui/skeleton";

import { useEnvironment } from "./use-environment";

export function EnvironmentCard() {
  const { info, refreshing, failed, refresh } = useEnvironment();
  const loading = !info && refreshing;
  const fields = [
    { label: "默认登录 Shell", value: info?.loginShell },
    { label: "用户主目录", value: info?.homeDir },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>运行环境</CardTitle>
        <CardAction>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={refreshing}
            onClick={() => void refresh()}
          >
            {refreshing ? "刷新中…" : failed ? "重试" : "刷新"}
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="@container flex flex-col gap-4">
        {failed && (
          <Alert variant="destructive">
            <AlertDescription>
              {info
                ? "刷新失败，当前显示上次读取的环境。"
                : "暂时无法读取运行环境。"}
            </AlertDescription>
          </Alert>
        )}
        {(info || loading) && (
          <ItemGroup
            className="grid gap-4 @xs:grid-cols-2"
            aria-label="运行环境"
            aria-live="polite"
            aria-busy={refreshing}
          >
            {fields.map(({ label, value }) => (
              <Item key={label} className="min-w-0 p-0" role="listitem">
                <ItemContent className="min-w-0">
                  <ItemDescription>{label}</ItemDescription>
                  {loading ? (
                    <Skeleton className="h-5 w-24" />
                  ) : value ? (
                    <code className="wrap-anywhere">{value}</code>
                  ) : (
                    <div>暂不可用</div>
                  )}
                </ItemContent>
              </Item>
            ))}
          </ItemGroup>
        )}
      </CardContent>
    </Card>
  );
}
