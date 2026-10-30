import { useRef, useState } from "react"
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
import {
  draftFromProxy,
  draftToProxy,
  isValidPort,
  proxySummary,
  type ProxyDraft,
} from "@/features/settings/model"
import { IpcClientError } from "@/lib/tauri-ipc"
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
  const [draft, setDraft] = useState<ProxyDraft>(() =>
    draftFromProxy(snapshot.proxy)
  )
  const [submitted, setSubmitted] = useState(false)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveNotice, setSaveNotice] = useState<string | null>(null)
  const [testError, setTestError] = useState<string | null>(null)
  const [testResult, setTestResult] = useState<ProxyTestResult | null>(null)
  const draftVersion = useRef(0)
  const testGeneration = useRef(0)

  const hostInvalid = draft.mode === "manual" && draft.host.trim() === ""
  const portInvalid = draft.mode === "manual" && !isValidPort(draft.port)
  const proxy = draftToProxy(draft)

  const updateDraft = (patch: Partial<ProxyDraft>) => {
    draftVersion.current += 1
    testGeneration.current += 1
    setDraft((current) => ({ ...current, ...patch }))
    setSaveError(null)
    setSaveNotice(null)
    setTestError(null)
    setTestResult(null)
    setTesting(false)
  }

  const save = async () => {
    setSubmitted(true)
    setSaveError(null)
    setSaveNotice(null)
    if (!proxy) {
      setSaveError("请填写有效的代理主机和 1 到 65535 之间的端口。")
      return
    }

    const submittedDraftVersion = draftVersion.current
    setSaving(true)
    try {
      const next = await onSave(snapshot.revision, proxy)
      if (draftVersion.current === submittedDraftVersion) {
        setDraft(draftFromProxy(next.proxy))
        setSaveNotice("代理设置已保存，将用于之后开始的下载任务。")
      } else {
        setSaveNotice("已保存提交时的代理设置；当前表单修改尚未保存。")
      }
    } catch (error) {
      setSaveError(formatIpcError(error))
    } finally {
      setSaving(false)
    }
  }

  const test = async () => {
    setSubmitted(true)
    setTestError(null)
    setTestResult(null)
    if (!proxy) {
      setTestError("请先填写有效的代理主机和端口。")
      return
    }

    const requestId = ++testGeneration.current
    setTesting(true)
    try {
      const result = await onTest(proxy)
      if (testGeneration.current === requestId) {
        setTestResult(result)
      }
    } catch (error) {
      if (testGeneration.current === requestId) {
        setTestError(formatIpcError(error))
      }
    } finally {
      if (testGeneration.current === requestId) {
        setTesting(false)
      }
    }
  }

  return (
    <Card className="max-w-3xl">
      <CardHeader>
        <CardTitle>下载代理</CardTitle>
        <CardDescription>
          DevBox
          后续的版本目录查询和工具下载会使用这里的网络设置。当前系统工具检测是本地操作，不经过代理。
        </CardDescription>
        <CardAction>
          <Badge variant="secondary">{proxySummary(snapshot.proxy)}</Badge>
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
            <FieldDescription>
              {proxyModeDescription(draft.mode)}
            </FieldDescription>
          </Field>

          {draft.mode === "manual" ? (
            <>
              <Field>
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
                <FieldDescription>
                  SOCKS5H 使用远端 DNS，适合代理服务器负责域名解析的场景。
                </FieldDescription>
              </Field>

              <Field data-invalid={submitted && hostInvalid}>
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
                <FieldDescription>
                  只填写主机名或 IP，不要包含协议和端口。
                </FieldDescription>
                {submitted && hostInvalid ? (
                  <FieldError>代理主机不能为空。</FieldError>
                ) : null}
              </Field>

              <Field data-invalid={submitted && portInvalid}>
                <FieldLabel htmlFor="proxy-port">代理端口</FieldLabel>
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
                <FieldDescription>端口范围为 1 到 65535。</FieldDescription>
                {submitted && portInvalid ? (
                  <FieldError>请输入有效的代理端口。</FieldError>
                ) : null}
              </Field>

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
                <FieldDescription>
                  可选，多个条目用逗号分隔。支持主机名、IP 和 CIDR，使用 *
                  表示全部绕过代理。
                </FieldDescription>
              </Field>
            </>
          ) : null}
        </FieldGroup>

        {saveError ? (
          <Alert variant="destructive">
            <TriangleAlertIcon />
            <AlertTitle>无法保存代理设置</AlertTitle>
            <AlertDescription>{saveError}</AlertDescription>
          </Alert>
        ) : null}

        {saveNotice ? (
          <Alert>
            <CircleCheckIcon />
            <AlertTitle>代理设置已保存</AlertTitle>
            <AlertDescription>{saveNotice}</AlertDescription>
          </Alert>
        ) : null}

        {testError ? (
          <Alert variant="destructive">
            <TriangleAlertIcon />
            <AlertTitle>代理连接失败</AlertTitle>
            <AlertDescription>{testError}</AlertDescription>
          </Alert>
        ) : null}

        {testResult ? (
          <Alert>
            <CircleCheckIcon />
            <AlertTitle>代理连接正常</AlertTitle>
            <AlertDescription>
              访问 GitHub 用时 {testResult.latencyMs} ms，HTTP 状态码{" "}
              {testResult.httpStatus}。
            </AlertDescription>
          </Alert>
        ) : null}
      </CardContent>
      <CardFooter className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-xs text-muted-foreground">
          设置由 Rust 校验并写入
          ~/.devbox/config.json，已有下载任务不会中途切换代理。
        </p>
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
            {testing ? "正在测试" : "测试当前输入"}
          </Button>
          <Button onClick={() => void save()} disabled={saving}>
            {saving ? (
              <Spinner data-icon="inline-start" />
            ) : (
              <SaveIcon data-icon="inline-start" />
            )}
            {saving ? "正在保存" : "保存设置"}
          </Button>
        </div>
      </CardFooter>
    </Card>
  )
}

function proxyModeDescription(mode: ProxyDraft["mode"]): string {
  switch (mode) {
    case "system":
      return "读取操作系统静态 HTTP/HTTPS 代理；Linux 没有统一系统代理时沿用终端代理环境变量。"
    case "manual":
      return "手动代理模式不会继承终端里的 HTTP_PROXY 或 HTTPS_PROXY。"
    case "direct":
      return "直连模式不读取系统代理或终端代理环境变量。"
  }
}

function formatIpcError(error: unknown): string {
  if (error instanceof IpcClientError) {
    return `${error.code} · ${error.message}`
  }
  return "无法完成设置操作"
}
