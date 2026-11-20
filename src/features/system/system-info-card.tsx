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
  ItemTitle,
} from "@/components/ui/item";
import { Skeleton } from "@/components/ui/skeleton";
import { formatBytes } from "@/features/system/formatters";
import { useSystemInfo } from "@/features/system/use-system-info";

export function SystemInfoCard() {
  const state = useSystemInfo();
  const systemInfo = state.snapshot?.info;
  const busy = state.pending !== null;
  const loading = !state.snapshot && busy;
  const pendingLabel = state.pending === "loading" ? "读取中…" : "刷新中…";

  const isAppleSilicon =
    systemInfo?.os === "darwin" && systemInfo.cpuModel.startsWith("Apple ");
  const graphics = systemInfo?.gpus.map(({ model }) => model).join("、");
  const memory = formatBytes(systemInfo?.memoryTotalBytes);
  const disk = formatBytes(systemInfo?.diskTotalBytes);

  const items = [
    {
      label: "操作系统",
      value: [systemInfo?.osName, systemInfo?.osVersion].filter(Boolean).join(" "),
    },
    {
      label: isAppleSilicon ? "芯片配置" : "处理器",
      value: systemInfo?.cpuModel,
    },
    ...(systemInfo && !isAppleSilicon
      ? [{ label: "图形处理器", value: graphics }]
      : []),
    {
      label: "内存与存储",
      value: [
        memory ?? "内存容量暂不可用",
        disk ?? "硬盘容量暂不可用",
      ].join(" · "),
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>系统信息</CardTitle>
        <CardAction>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={busy}
            onClick={() => void state.refresh()}
          >
            {busy ? pendingLabel : state.error ? "重试" : "刷新"}
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="@container flex flex-col gap-4">
        {state.error && (
          <Alert variant="destructive">
            <AlertDescription>
              {state.snapshot
                ? "刷新失败，当前显示上次采集的信息。"
                : state.error}
            </AlertDescription>
          </Alert>
        )}
        {(state.snapshot || loading) && (
          <ItemGroup
            className="grid @xs:grid-cols-2 @lg:grid-cols-3"
            aria-label={loading ? "正在读取系统信息" : "系统信息"}
            aria-live="polite"
            aria-busy={busy}
          >
            {items.map(({ label, value }) => (
              <Item key={label} className="min-w-0 p-0" role="listitem">
                <ItemContent className="min-w-0">
                  <ItemDescription>{label}</ItemDescription>
                  <ItemTitle title={value || undefined}>
                    {loading ? (
                      <Skeleton className="h-5 w-24" />
                    ) : (
                      value || "暂不可用"
                    )}
                  </ItemTitle>
                </ItemContent>
              </Item>
            ))}
          </ItemGroup>
        )}
      </CardContent>
    </Card>
  );
}
