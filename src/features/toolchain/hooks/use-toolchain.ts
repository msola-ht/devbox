import { listTools, refreshTools } from "@/features/toolchain/api"
import {
  createSnapshotStore,
  useSnapshotStore,
} from "@/hooks/use-snapshot-store"

const toolchainStore = createSnapshotStore({
  query: listTools,
  refresh: refreshTools,
})

export function useToolchain() {
  return useSnapshotStore(toolchainStore)
}
