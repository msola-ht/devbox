import { getSettings, testProxy, updateSettings } from "@/features/settings/api"
import type { ProxySettings } from "@/features/settings/contracts"
import {
  createSnapshotStore,
  useSnapshotStore,
} from "@/hooks/use-snapshot-store"
import { IpcClientError } from "@/lib/tauri-ipc"

const settingsStore = createSnapshotStore({ query: getSettings })

export function useSettings() {
  const snapshot = useSnapshotStore(settingsStore)

  return {
    ...snapshot,
    async update(expectedRevision: string, proxy: ProxySettings) {
      try {
        const next = await updateSettings({ expectedRevision, proxy })
        settingsStore.apply(next)
        return next
      } catch (error) {
        if (
          error instanceof IpcClientError &&
          error.code === "SETTINGS_CONFLICT"
        ) {
          await settingsStore.refresh()
        }
        throw error
      }
    },
    testProxy,
  }
}
