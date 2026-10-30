import type { ToolListSnapshot } from "@/features/toolchain/contracts"
import { invokeCommand } from "@/lib/tauri-ipc"

export function listTools(): Promise<ToolListSnapshot> {
  return invokeCommand<ToolListSnapshot>("list_tools")
}

export function refreshTools(): Promise<ToolListSnapshot> {
  return invokeCommand<ToolListSnapshot>("refresh_tools")
}
