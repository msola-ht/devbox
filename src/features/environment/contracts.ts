export interface EnvironmentSnapshot {
  bootId: string
  revision: string
  generatedAt: string
  os: string
  osVersion: string | null
  osBuild: string | null
  arch: string
  homeDir: string
  devboxRoot: string
  devboxReady: boolean
  devboxError: string | null
  memoryTotalBytes: number | null
  diskTotalBytes: number | null
  gpuName: string | null
  gpuCoreCount: number | null
}
