import { getEnvironment } from "@/features/environment/api"
import {
  createSnapshotStore,
  useSnapshotStore,
} from "@/hooks/use-snapshot-store"

const environmentStore = createSnapshotStore({ query: getEnvironment })

export function useEnvironment() {
  return useSnapshotStore(environmentStore)
}
