import type {
  ProxySettings,
  ProxyTestResult,
  SettingsSnapshot,
  UpdateSettingsRequest,
} from "@/features/settings/contracts"
import { invokeCommand } from "@/lib/tauri-ipc"

export function getSettings(): Promise<SettingsSnapshot> {
  return invokeCommand<SettingsSnapshot>("get_settings")
}

export function updateSettings(
  request: UpdateSettingsRequest
): Promise<SettingsSnapshot> {
  return invokeCommand<SettingsSnapshot>("update_settings", { request })
}

export function testProxy(proxy: ProxySettings): Promise<ProxyTestResult> {
  return invokeCommand<ProxyTestResult>("test_proxy", { request: { proxy } })
}
