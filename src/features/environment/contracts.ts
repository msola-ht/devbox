export interface EnvironmentSnapshot {
  bootId: string
  revision: string
  generatedAt: string
  os: string
  arch: string
  homeDir: string
  devboxRoot: string
  devboxReady: boolean
  devboxError: string | null
}
