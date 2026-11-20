export function formatBytes(bytes: number | null | undefined) {
  if (bytes == null) return undefined;

  const units = ["B", "KiB", "MiB", "GiB", "TiB", "PiB"];
  let value = bytes;
  let unit = 0;

  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit += 1;
  }

  return `${value.toLocaleString("zh-CN", { maximumFractionDigits: 1 })} ${units[unit]}`;
}
