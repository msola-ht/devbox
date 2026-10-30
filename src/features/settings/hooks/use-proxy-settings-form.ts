import { useRef, useState } from "react"

import type {
  ProxySettings,
  ProxyTestResult,
  SettingsSnapshot,
} from "@/features/settings/contracts"
import {
  draftFromProxy,
  draftToProxy,
  isValidPort,
  type ProxyDraft,
} from "@/features/settings/model"
import { IpcClientError } from "@/lib/tauri-ipc"
import { toast } from "@/components/ui/toast"

interface ProxySettingsFormArgs {
  snapshot: SettingsSnapshot
  onSave: (
    expectedRevision: string,
    proxy: ProxySettings
  ) => Promise<SettingsSnapshot>
  onTest: (proxy: ProxySettings) => Promise<ProxyTestResult>
}

export function useProxySettingsForm({
  snapshot,
  onSave,
  onTest,
}: ProxySettingsFormArgs) {
  const [draft, setDraft] = useState<ProxyDraft>(() =>
    draftFromProxy(snapshot.proxy)
  )
  const [submitted, setSubmitted] = useState(false)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveNotice, setSaveNotice] = useState<string | null>(null)
  const [testError, setTestError] = useState<string | null>(null)
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
    setTesting(false)
  }

  const save = async () => {
    setSubmitted(true)
    setSaveError(null)
    setSaveNotice(null)
    if (!proxy) {
      setSaveError("请填写有效的代理地址和端口。")
      return
    }

    const submittedDraftVersion = draftVersion.current
    setSaving(true)
    try {
      const next = await onSave(snapshot.revision, proxy)
      if (draftVersion.current === submittedDraftVersion) {
        setDraft(draftFromProxy(next.proxy))
        setSaveNotice("设置已保存。")
      } else {
        setSaveNotice("已保存提交的版本，当前修改尚未保存。")
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
    if (!proxy) {
      setTestError("请填写有效的代理地址和端口。")
      return
    }

    const requestId = ++testGeneration.current
    setTesting(true)
    try {
      const result = await onTest(proxy)
      if (testGeneration.current === requestId) {
        toast.add({
          type: "success",
          title: "连接正常",
          description: `GitHub · ${result.latencyMs} ms · HTTP ${result.httpStatus}`,
        })
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

  return {
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
  }
}

function formatIpcError(error: unknown): string {
  if (error instanceof IpcClientError) {
    return `${error.code} · ${error.message}`
  }
  return "无法完成设置操作"
}
