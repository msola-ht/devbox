import type { EnvironmentSnapshot } from "@/features/environment/contracts"
import { invokeCommand } from "@/lib/tauri-ipc"

export function getEnvironment(): Promise<EnvironmentSnapshot> {
  return invokeCommand<EnvironmentSnapshot>("get_environment")
}
