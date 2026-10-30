import { getEnvironment, refreshEnvironment } from "@/features/environment/api"
import {
  createSnapshotStore,
  useSnapshotStore,
} from "@/hooks/use-snapshot-store"

const environmentStore = createSnapshotStore({
  query: getEnvironment,
  refresh: refreshEnvironment,
})

export function useEnvironment() {
  return useSnapshotStore(environmentStore)
}
