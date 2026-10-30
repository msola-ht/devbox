import {
  CircleCheckIcon,
  PlugZapIcon,
  SaveIcon,
  TriangleAlertIcon,
} from "lucide-react"

import type {
  ProxySettings,
  ProxyTestResult,
  SettingsSnapshot,
} from "@/features/settings/contracts"
import { SystemProxySummary } from "@/features/settings/components/system-proxy-summary"
import { useProxySettingsForm } from "@/features/settings/hooks/use-proxy-settings-form"
import { proxySummary } from "@/features/settings/model"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldTitle,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"

interface ProxySettingsCardProps {
  snapshot: SettingsSnapshot
  onSave: (
    expectedRevision: string,
    proxy: ProxySettings
  ) => Promise<SettingsSnapshot>
  onTest: (proxy: ProxySettings) => Promise<ProxyTestResult>
}

export function ProxySettingsCard({
  snapshot,
  onSave,
  onTest,
}: ProxySettingsCardProps) {
  const {
    draft,
    submitted,
    saving,
    testing,
    saveError,
    saveNotice,
    testError,
    hostInvalid,
    portInvalid,
    updateDraft,
    save,
    test,
  } = useProxySettingsForm({ snapshot, onSave, onTest })

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>下载代理</CardTitle>
        <CardDescription>用于版本查询和工具下载。</CardDescription>
        <CardAction>
          <Badge variant="secondary" className="max-w-64 truncate">
            {proxySummary(snapshot.proxy)}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-col gap-5">
        <FieldGroup>
          <Field>
            <FieldTitle id="proxy-mode-label">连接方式</FieldTitle>
            <ToggleGroup
              aria-labelledby="proxy-mode-label"
              value={[draft.mode]}
              onValueChange={(value) => {
                const mode = value[0]
                if (
                  mode === "system" ||
                  mode === "direct" ||
                  mode === "manual"
                ) {
                  updateDraft({ mode })
                }
              }}
              variant="outline"
              spacing={2}
            >
              <ToggleGroupItem value="system">系统代理</ToggleGroupItem>
              <ToggleGroupItem value="manual">手动代理</ToggleGroupItem>
              <ToggleGroupItem value="direct">直连</ToggleGroupItem>
            </ToggleGroup>
            {draft.mode === "system" ? (
              <SystemProxySummary inspection={snapshot.systemProxy} />
            ) : null}
          </Field>

          {draft.mode === "manual" ? (
            <>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
                <Field className="sm:w-auto">
                  <FieldTitle id="proxy-protocol-label">代理协议</FieldTitle>
                  <ToggleGroup
                    aria-labelledby="proxy-protocol-label"
                    value={[draft.protocol]}
                    onValueChange={(value) => {
                      const protocol = value[0]
                      if (protocol === "http" || protocol === "socks5h") {
                        updateDraft({ protocol })
                      }
                    }}
                    variant="outline"
                    spacing={2}
                  >
                    <ToggleGroupItem value="http">HTTP</ToggleGroupItem>
                    <ToggleGroupItem value="socks5h">SOCKS5H</ToggleGroupItem>
                  </ToggleGroup>
                </Field>

                <Field
                  className="sm:min-w-0 sm:flex-1"
                  data-invalid={submitted && hostInvalid}
                >
                  <FieldLabel htmlFor="proxy-host">代理主机</FieldLabel>
                  <Input
                    id="proxy-host"
                    value={draft.host}
                    onChange={(event) =>
                      updateDraft({ host: event.target.value })
                    }
                    placeholder="127.0.0.1"
                    aria-invalid={submitted && hostInvalid}
                    autoComplete="off"
                    spellCheck={false}
                  />
                  {submitted && hostInvalid ? (
                    <FieldError>请填写代理主机。</FieldError>
                  ) : null}
                </Field>

                <Field
                  className="sm:w-24"
                  data-invalid={submitted && portInvalid}
                >
                  <FieldLabel htmlFor="proxy-port">端口</FieldLabel>
                  <Input
                    id="proxy-port"
                    value={draft.port}
                    onChange={(event) =>
                      updateDraft({ port: event.target.value })
                    }
                    placeholder="7890"
                    inputMode="numeric"
                    aria-invalid={submitted && portInvalid}
                    autoComplete="off"
                  />
                  {submitted && portInvalid ? (
                    <FieldError>端口无效</FieldError>
                  ) : null}
                </Field>
              </div>

              <Field>
                <FieldLabel htmlFor="proxy-bypass">绕过代理</FieldLabel>
                <Input
                  id="proxy-bypass"
                  value={draft.bypass}
                  onChange={(event) =>
                    updateDraft({ bypass: event.target.value })
                  }
                  placeholder="localhost, 127.0.0.1, ::1"
                  autoComplete="off"
                  spellCheck={false}
                />
                <FieldDescription>多个条目用逗号分隔。</FieldDescription>
              </Field>
            </>
          ) : null}
        </FieldGroup>

        {saveError ? (
          <Alert variant="destructive">
            <TriangleAlertIcon />
            <AlertTitle>保存失败</AlertTitle>
            <AlertDescription>{saveError}</AlertDescription>
          </Alert>
        ) : null}

        {saveNotice ? (
          <Alert>
            <CircleCheckIcon />
            <AlertTitle>已保存</AlertTitle>
            <AlertDescription>{saveNotice}</AlertDescription>
          </Alert>
        ) : null}

        {testError ? (
          <Alert variant="destructive">
            <TriangleAlertIcon />
            <AlertTitle>连接失败</AlertTitle>
            <AlertDescription>{testError}</AlertDescription>
          </Alert>
        ) : null}
      </CardContent>
      <CardFooter className="flex flex-wrap items-center justify-end gap-3">
        <div className="flex flex-wrap gap-2">
          <Button
            variant="outline"
            onClick={() => void test()}
            disabled={testing}
          >
            {testing ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <PlugZapIcon data-icon="inline-start" />
            )}
            {testing ? "测试中" : "测试"}
          </Button>
          <Button onClick={() => void save()} disabled={saving}>
            {saving ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <SaveIcon data-icon="inline-start" />
            )}
            {saving ? "保存中" : "保存"}
          </Button>
        </div>
      </CardFooter>
    </Card>
  )
}
