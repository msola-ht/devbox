import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Card,
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
import type { EnvironmentInfo } from "@/mygo";

interface EnvironmentCardProps {
  info: EnvironmentInfo | null;
  busy: boolean;
  failed: boolean;
}

export function EnvironmentCard({ info, busy, failed }: EnvironmentCardProps) {
  const loading = !info && busy;
  const fields = [
    { label: "默认登录 Shell", value: info?.loginShell },
    { label: "用户主目录", value: info?.homeDir },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>运行环境</CardTitle>
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
            aria-busy={busy}
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
