import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Card,
  CardContent,
  CardDescription,
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
  const systemInfo = state.status === "ready" ? state.data : null;
  const loading = state.status === "loading";

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
        memory ? `${memory} 内存` : "内存容量暂不可用",
        disk ? `${disk} 硬盘` : "硬盘容量暂不可用",
      ].join(" · "),
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>系统信息</CardTitle>
        <CardDescription>这台设备的系统与硬件配置。</CardDescription>
      </CardHeader>
      <CardContent>
        {state.status === "error" ? (
          <Alert variant="destructive">
            <AlertDescription>暂时无法读取系统信息。</AlertDescription>
          </Alert>
        ) : (
          <ItemGroup
            className="grid sm:grid-cols-3"
            aria-label={loading ? "正在读取系统信息" : "系统信息"}
            aria-live="polite"
            aria-busy={loading}
          >
            {items.map(({ label, value }) => (
              <Item key={label} className="p-0" role="listitem">
                <ItemContent>
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
