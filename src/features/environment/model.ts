const byteFormatter = new Intl.NumberFormat("zh-CN", {
  maximumFractionDigits: 1,
})

export function displayOs(os: string) {
  if (os === "macos") {
    return "macOS"
  }
  if (os === "windows") {
    return "Windows"
  }
  if (os === "linux") {
    return "Linux"
  }
  return os
}

export function displayOsVersion(version: string | null, build: string | null) {
  if (!version) {
    return "未知"
  }

  return build ? `${version} (${build})` : version
}

export function formatBytes(bytes: number | null) {
  if (bytes === null || !Number.isFinite(bytes) || bytes < 0) {
    return "未知"
  }

  const units = ["B", "KB", "MB", "GB", "TB", "PB"]
  let value = bytes
  let unitIndex = 0

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }

  return `${byteFormatter.format(value)} ${units[unitIndex]}`
}

export function displayGpu(name: string | null, cores: number | null) {
  if (!name) {
    return "未知"
  }

  return cores === null ? name : `${name}（${cores} 核）`
}
