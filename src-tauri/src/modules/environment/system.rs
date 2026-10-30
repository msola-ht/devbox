use std::path::Path;
use std::time::Duration;

use crate::infrastructure::process::probe_version;

const PROBE_TIMEOUT: Duration = Duration::from_secs(2);

#[derive(Clone, Debug, Default)]
pub struct SystemProfile {
  pub os_version: Option<String>,
  pub os_build: Option<String>,
  pub memory_total_bytes: Option<u64>,
  pub disk_total_bytes: Option<u64>,
  pub gpu_name: Option<String>,
  pub gpu_core_count: Option<u32>,
}

pub async fn detect_system_profile(home_dir: &Path) -> SystemProfile {
  let home_dir = home_dir.to_path_buf();
  let os = tokio::spawn(detect_os_version());
  let gpu = tokio::spawn(detect_gpu());
  let memory = tokio::spawn(detect_memory_total());
  let disk = tokio::spawn(async move { detect_disk_total(&home_dir).await });

  let (os_version, os_build) = os.await.unwrap_or((None, None));
  let memory_total_bytes = memory.await.ok().flatten();
  let disk_total_bytes = disk.await.ok().flatten();
  let (gpu_name, gpu_core_count) = gpu.await.unwrap_or((None, None));

  SystemProfile {
    os_version,
    os_build,
    memory_total_bytes,
    disk_total_bytes,
    gpu_name,
    gpu_core_count,
  }
}

async fn probe(program: &str, args: &[&str]) -> Option<String> {
  match probe_version(Path::new(program), args, PROBE_TIMEOUT).await {
    Ok(output) => Some(output),
    Err(error) => {
      log::debug!("系统信息命令失败 {program}: {error}");
      None
    }
  }
}

#[cfg(target_os = "macos")]
async fn detect_os_version() -> (Option<String>, Option<String>) {
  let Some(output) = probe("/usr/bin/sw_vers", &[]).await else {
    return (None, None);
  };
  let mut version = None;
  let mut build = None;

  for line in output.lines() {
    let Some((key, value)) = line.split_once(':') else {
      continue;
    };
    let value = value.trim();
    if value.is_empty() {
      continue;
    }

    match key.trim() {
      "ProductVersion" => version = Some(value.to_string()),
      "BuildVersion" => build = Some(value.to_string()),
      _ => {}
    }
  }

  (version, build)
}

#[cfg(target_os = "linux")]
async fn detect_os_version() -> (Option<String>, Option<String>) {
  let Ok(contents) = std::fs::read_to_string("/etc/os-release") else {
    return (None, None);
  };
  let version = contents.lines().find_map(|line| {
    let value = line.strip_prefix("VERSION_ID=")?;
    let value = value.trim().trim_matches('"');
    (!value.is_empty()).then(|| value.to_string())
  });

  (version, None)
}

#[cfg(windows)]
async fn detect_os_version() -> (Option<String>, Option<String>) {
  let Some(output) = probe("cmd.exe", &["/C", "ver"]).await else {
    return (None, None);
  };
  let version = output
    .split('[')
    .nth(1)
    .and_then(|value| value.split(']').next())
    .and_then(|value| {
      let start = value.find(|character: char| character.is_ascii_digit())?;
      let end = value.rfind(|character: char| character.is_ascii_digit())?;
      Some(value[start..=end].to_string())
    });

  (version, None)
}

#[cfg(not(any(target_os = "macos", target_os = "linux", windows)))]
async fn detect_os_version() -> (Option<String>, Option<String>) {
  (None, None)
}

#[cfg(target_os = "macos")]
async fn detect_memory_total() -> Option<u64> {
  probe("/usr/sbin/sysctl", &["-n", "hw.memsize"])
    .await
    .and_then(|output| output.trim().parse::<u64>().ok())
}

#[cfg(target_os = "linux")]
async fn detect_memory_total() -> Option<u64> {
  let contents = std::fs::read_to_string("/proc/meminfo").ok()?;
  let value = contents
    .lines()
    .find(|line| line.starts_with("MemTotal:"))?
    .strip_prefix("MemTotal:")?
    .split_whitespace()
    .next()?
    .parse::<u64>()
    .ok()?;

  value.checked_mul(1024)
}

#[cfg(not(any(target_os = "macos", target_os = "linux")))]
async fn detect_memory_total() -> Option<u64> {
  None
}

#[cfg(unix)]
async fn detect_disk_total(home_dir: &Path) -> Option<u64> {
  let home_dir = home_dir.to_string_lossy();
  let output = probe("/bin/df", &["-P", "-k", home_dir.as_ref()]).await?;

  output.lines().find_map(|line| {
    let mut fields = line.split_whitespace();
    fields.next()?;
    let total_kib = fields.next()?.parse::<u64>().ok()?;
    total_kib.checked_mul(1024)
  })
}

#[cfg(not(unix))]
async fn detect_disk_total(_home_dir: &Path) -> Option<u64> {
  None
}

#[cfg(target_os = "macos")]
async fn detect_gpu() -> (Option<String>, Option<u32>) {
  let Some(output) = probe(
    "/usr/sbin/ioreg",
    &["-rd1", "-c", "IOAccelerator", "-w", "0"],
  )
  .await
  else {
    return (None, None);
  };

  let mut names = Vec::new();
  let mut core_count = None;
  for line in output.lines() {
    let Some((key, value)) = line.split_once('=') else {
      continue;
    };
    let key = key.trim().trim_matches('"');
    let value = value.trim().trim_matches('"');

    if key == "model"
      && !value.is_empty()
      && !names.iter().any(|name: &String| name.as_str() == value)
    {
      names.push(value.to_string());
    } else if key == "gpu-core-count" && core_count.is_none() {
      core_count = value.parse::<u32>().ok();
    }
  }

  let name = (!names.is_empty()).then(|| names.join(" / "));
  (name, core_count)
}

#[cfg(not(target_os = "macos"))]
async fn detect_gpu() -> (Option<String>, Option<u32>) {
  (None, None)
}
