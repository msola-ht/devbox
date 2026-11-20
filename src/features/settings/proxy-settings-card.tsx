import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Field, FieldDescription, FieldGroup, FieldLabel, FieldSet, FieldTitle,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { addressChoices, proxyModes, useProxySettings } from "@/features/settings/use-proxy-settings";
import type { SystemProxyInfo, SystemProxyKind, SystemRoute } from "@/mygo";

const systemProxyNames: Record<SystemProxyKind, string> = {
  http: "HTTP",
  https: "HTTPS",
  socks: "SOCKS",
  auto_config: "自动配置",
  auto_discovery: "自动发现",
};

function isAutomatic(kind: SystemProxyKind) {
  return kind === "auto_config" || kind === "auto_discovery";
}

function routeLabel(route: SystemRoute) {
  switch (route.status) {
    case "direct": return "直连";
    case "error": return "解析失败";
    case "proxy": return `${route.protocol}://${route.address}`;
  }
}

function systemDisplayEntries(info: SystemProxyInfo) {
  const rows = info.settings.map(({ kind, address }) => ({
    name: systemProxyNames[kind],
    address: isAutomatic(kind) ? "已开启" : address || "地址不可用",
  }));
  const showAllRoutes = !info.settingsAvailable || info.settings.some(({ kind }) => isAutomatic(kind));
  const socks = info.settings.find(({ kind }) => kind === "socks");

  for (const [kind, route] of [["http", info.http], ["https", info.https]] as const) {
    const configured = info.settings.find((entry) => entry.kind === kind) ?? socks;
    const duplicate = configured
      ? route.status === "proxy" &&
        route.protocol === (configured.kind === "socks" ? "socks5" : "http") &&
        route.address === configured.address
      : route.status === "direct";
    if (showAllRoutes || !duplicate) {
      rows.push({ name: `应用 ${systemProxyNames[kind]}`, address: routeLabel(route) });
    }
  }
  return rows;
}

function SystemProxyStatus({ info, error }: { info: SystemProxyInfo | null; error: string | null }) {
  const rows = info ? systemDisplayEntries(info) : [];
  return (
    <>
      {error && <FieldDescription role="alert">{error}</FieldDescription>}
      {rows.length > 0 ? (
        <FieldGroup className="grid grid-cols-3" aria-live="polite">
          {rows.map(({ name, address }) => (
            <Field key={name}>
              <FieldTitle>{name}</FieldTitle>
              <FieldDescription>{address}</FieldDescription>
            </Field>
          ))}
        </FieldGroup>
      ) : !error && <FieldDescription role="status">{info ? "未设置" : "读取中…"}</FieldDescription>}
    </>
  );
}

export function ProxySettingsCard() {
  const state = useProxySettings();
  const { action, settings } = state;
  const draft = settings.status === "ready" ? settings.draft : null;

  return (
    <form onSubmit={(event) => {
      event.preventDefault();
      void state.save();
    }}>
      <Card>
        <CardHeader>
          <CardTitle>网络代理</CardTitle>
          <CardDescription>应用联网设置</CardDescription>
        </CardHeader>
        <CardContent>
          {settings.status === "loading" && (
            <Alert role="status">
              <AlertDescription>读取中…</AlertDescription>
            </Alert>
          )}
          {settings.status === "failed" && (
            <Alert variant="destructive"><AlertDescription>{settings.error}</AlertDescription></Alert>
          )}
          {settings.status === "invalid" && (
            <Alert variant="destructive">
              <AlertTitle>代理设置文件无效或版本不兼容</AlertTitle>
              <AlertDescription>恢复默认设置将替换当前代理设置文件。</AlertDescription>
            </Alert>
          )}
          {draft && (
            <FieldSet disabled={action !== null}>
              <FieldGroup>
                <Field>
                  <FieldTitle id="proxy-mode-label">代理模式</FieldTitle>
                  <ToggleGroup
                    aria-labelledby="proxy-mode-label"
                    variant="outline"
                    value={[draft.mode]}
                    onValueChange={([value]) => {
                      const mode = proxyModes.find((option) => option.value === value)?.value;
                      if (mode) state.change({ mode });
                    }}
                  >
                    {proxyModes.map(({ value, label }) => (
                      <ToggleGroupItem key={value} value={value}>{label}</ToggleGroupItem>
                    ))}
                  </ToggleGroup>
                </Field>
                {draft.mode === "system" && <SystemProxyStatus info={state.systemInfo} error={state.systemError} />}
                {draft.mode === "manual" && (
                  <>
                    <Field>
                      <FieldTitle id="proxy-address-label">代理地址</FieldTitle>
                      <ToggleGroup
                        aria-labelledby="proxy-address-label"
                        variant="outline"
                        value={[draft.choice]}
                        onValueChange={([value]) => {
                          const choice = addressChoices.find((option) => option.value === value)?.value;
                          if (choice) state.change({ choice });
                        }}
                      >
                        {addressChoices.map(({ value, label }) => (
                          <ToggleGroupItem key={value} value={value}>{label}</ToggleGroupItem>
                        ))}
                      </ToggleGroup>
                    </Field>
                    {draft.choice === "custom" && (
                      <Field>
                        <FieldLabel htmlFor="proxy-address">自定义地址</FieldLabel>
                        <Input
                          id="proxy-address"
                          value={draft.customAddress}
                          placeholder="127.0.0.1:7890"
                          onChange={(event) => state.change({ customAddress: event.target.value })}
                        />
                      </Field>
                    )}
                  </>
                )}
              </FieldGroup>
            </FieldSet>
          )}
        </CardContent>
        {settings.status === "failed" && (
          <CardFooter>
            <Button type="button" variant="outline" onClick={() => void state.load()}>重试</Button>
          </CardFooter>
        )}
        {settings.status === "invalid" && (
          <CardFooter>
            <Button
              type="button"
              variant="destructive"
              disabled={action === "resetting"}
              aria-busy={action === "resetting"}
              onClick={() => void state.reset()}
            >
              {action === "resetting" && <Spinner data-icon="inline-start" aria-hidden />}
              恢复默认设置
            </Button>
          </CardFooter>
        )}
        {draft && (
          <CardFooter>
            <Field orientation="horizontal">
              <Button type="submit" disabled={action !== null} aria-busy={action === "saving"}>
                保存设置
              </Button>
              {action === "checking" ? (
                <Button type="button" variant="outline" onClick={state.cancelCheck}>取消检查</Button>
              ) : (
                <Button type="button" variant="outline" disabled={action === "saving"} onClick={() => void state.check()}>
                  检查连接
                </Button>
              )}
            </Field>
          </CardFooter>
        )}
      </Card>
    </form>
  );
}
