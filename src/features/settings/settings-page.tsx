import { Settings2Icon } from "lucide-react";

import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

export function SettingsPage() {
  return (
    <Empty>
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Settings2Icon />
        </EmptyMedia>
        <EmptyTitle>设置</EmptyTitle>
        <EmptyDescription>暂无设置项。</EmptyDescription>
      </EmptyHeader>
    </Empty>
  );
}
