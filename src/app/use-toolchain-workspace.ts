import { useCallback } from "react"

import { useEnvironment } from "@/features/environment"
import { useToolchain } from "@/features/toolchain"

export function useToolchainWorkspace() {
  const environment = useEnvironment()
  const toolchain = useToolchain()
  const { refresh: refreshEnvironment } = environment
  const { refresh: refreshToolchain } = toolchain
  const loading =
    environment.state.status === "loading" ||
    toolchain.state.status === "loading"

  const refreshAll = useCallback(async () => {
    await Promise.all([refreshEnvironment(), refreshToolchain()])
  }, [refreshEnvironment, refreshToolchain])

  return {
    environment,
    toolchain,
    loading,
    refreshAll,
  }
}
