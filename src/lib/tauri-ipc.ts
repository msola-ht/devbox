import { invoke, type InvokeArgs } from "@tauri-apps/api/core"

export class IpcClientError extends Error {
  readonly code: string
  readonly details: unknown

  constructor(code: string, message: string, details?: unknown) {
    super(message)
    this.name = "IpcClientError"
    this.code = code
    this.details = details
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}

function normalizeIpcError(error: unknown): IpcClientError {
  if (error instanceof IpcClientError) {
    return error
  }

  if (isRecord(error)) {
    const code =
      typeof error.code === "string" ? error.code : "IPC_TRANSPORT_FAILED"
    const message =
      typeof error.message === "string" ? error.message : "无法完成本地请求"
    return new IpcClientError(code, message, error.details)
  }

  if (typeof error === "string") {
    return new IpcClientError("IPC_TRANSPORT_FAILED", error)
  }

  return new IpcClientError("IPC_TRANSPORT_FAILED", "无法完成本地请求")
}

export async function invokeCommand<T>(
  command: string,
  args?: InvokeArgs
): Promise<T> {
  try {
    if (typeof window === "undefined" || !("__TAURI_INTERNALS__" in window)) {
      throw new IpcClientError(
        "DESKTOP_UNAVAILABLE",
        "当前页面不在 DevBox 桌面应用中，无法读取本机状态"
      )
    }
    return await invoke<T>(command, args)
  } catch (error) {
    throw normalizeIpcError(error)
  }
}
