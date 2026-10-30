import { useCallback } from "react"

import { useEnvironment } from "@/features/environment"
import { selectVisibleTools, useToolchain } from "@/features/toolchain"

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
  const tools =
    hasCurrentToolSnapshot && toolchain.state.data
      ? selectVisibleTools(toolchain.state.data.tools)
      : []

  const refreshAll = useCallback(async () => {
    await refreshEnvironment()
    await refreshToolchain()
  }, [refreshEnvironment, refreshToolchain])

  return {
    environment,
    toolchain,
    tools,
    hasCurrentToolSnapshot,
    loading,
    refreshAll,
  }
}
