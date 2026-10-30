import { useCallback } from "react"

import { useEnvironment } from "@/features/environment"
import {
  selectSwitchableShells,
  selectVisibleTools,
  useToolchain,
} from "@/features/toolchain"

export function useToolchainWorkspace() {
  const environment = useEnvironment()
  const toolchain = useToolchain()
  const { refresh: refreshEnvironment } = environment
  const { refresh: refreshToolchain } = toolchain
  const loading =
    environment.state.status === "loading" ||
    toolchain.state.status === "loading"
  const hasCurrentToolSnapshot =
    environment.state.data !== null &&
    toolchain.state.data !== null &&
    environment.state.data.revision === toolchain.state.data.environmentRevision
  const snapshotTools =
    hasCurrentToolSnapshot && toolchain.state.data
      ? toolchain.state.data.tools
      : []
  const tools = selectVisibleTools(snapshotTools)
  const switchableShells = selectSwitchableShells(snapshotTools)

  const refreshAll = useCallback(async () => {
    await refreshEnvironment()
    await refreshToolchain()
  }, [refreshEnvironment, refreshToolchain])

  return {
    environment,
    toolchain,
    tools,
    switchableShells,
    hasCurrentToolSnapshot,
    loading,
    refreshAll,
  }
}
