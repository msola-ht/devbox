use std::ffi::OsStr;
use std::path::{Path, PathBuf};
use std::time::Duration;

use crate::infrastructure::process::probe_version;
use crate::modules::environment::ShellEnvironmentStatus;

const PROBE_TIMEOUT: Duration = Duration::from_secs(2);
const LOGIN_SHELL_TIMEOUT: Duration = Duration::from_secs(5);
const LOGIN_PATH_BEGIN: &str = "__DEVBOX_LOGIN_PATH_BEGIN__";
const LOGIN_PATH_END: &str = "__DEVBOX_LOGIN_PATH_END__";

#[derive(Clone, Debug, Default)]
pub struct SystemProfile {
  pub os_version: Option<String>,
  pub os_build: Option<String>,
  pub memory_total_bytes: Option<u64>,
  pub disk_total_bytes: Option<u64>,
  pub gpu_name: Option<String>,
  pub gpu_core_count: Option<u32>,
  pub default_shell: Option<String>,
  pub shell_environment_status: ShellEnvironmentStatus,
  pub login_shell_path: Vec<PathBuf>,
}

pub async fn detect_system_profile(home_dir: &Path) -> SystemProfile {
  let home_dir = home_dir.to_path_buf();
  let os = tokio::spawn(detect_os_version());
  let gpu = tokio::spawn(detect_gpu());
  let memory = tokio::spawn(detect_memory_total());
  let shell_home = home_dir.clone();
  let default_shell = tokio::spawn(detect_default_shell(shell_home));
  let disk = tokio::spawn(async move { detect_disk_total(&home_dir).await });

  let (os_version, os_build) = os.await.unwrap_or((None, None));
  let memory_total_bytes = memory.await.ok().flatten();
  let disk_total_bytes = disk.await.ok().flatten();
  let (gpu_name, gpu_core_count) = gpu.await.unwrap_or((None, None));
  let default_shell = default_shell.await.ok().flatten();
  let (login_shell_path, shell_environment_status) = match default_shell.as_deref() {
    Some(shell) => detect_login_shell_path(Path::new(shell)).await,
    None => (Vec::new(), ShellEnvironmentStatus::DefaultShellUnavailable),
  };

  SystemProfile {
    os_version,
    os_build,
    memory_total_bytes,
    disk_total_bytes,
    gpu_name,
    gpu_core_count,
    default_shell,
    shell_environment_status,
    login_shell_path,
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

async fn detect_login_shell_path(shell: &Path) -> (Vec<PathBuf>, ShellEnvironmentStatus) {
  let Some(shell_name) = shell.file_name().and_then(OsStr::to_str) else {
    return (Vec::new(), ShellEnvironmentStatus::DefaultShellUnavailable);
  };
  if !shell.is_absolute() || !matches!(shell_name.trim_end_matches(".exe"), "bash" | "zsh") {
    return (Vec::new(), ShellEnvironmentStatus::UnsupportedShell);
  }

  let script = format!("printf '\\n{LOGIN_PATH_BEGIN}%s{LOGIN_PATH_END}\\n' \"$PATH\"");
  let output = match probe_version(shell, &["-l", "-c", script.as_str()], LOGIN_SHELL_TIMEOUT).await
  {
    Ok(output) => output,
    Err(_) => {
      log::warn!("无法读取默认 Shell 登录环境 {}", shell.display());
      return (Vec::new(), ShellEnvironmentStatus::LoginPathUnavailable);
    }
  };

  let Some(value_start) = output.find(LOGIN_PATH_BEGIN) else {
    return (Vec::new(), ShellEnvironmentStatus::LoginPathUnavailable);
  };
  let value_start = value_start + LOGIN_PATH_BEGIN.len();
  let Some(value_end) = output[value_start..].find(LOGIN_PATH_END) else {
    return (Vec::new(), ShellEnvironmentStatus::LoginPathUnavailable);
  };
  let value = &output[value_start..value_start + value_end];

  let login_shell_path = std::env::split_paths(OsStr::new(value))
    .filter(|path| path.is_absolute())
    .collect::<Vec<_>>();
  let status = if login_shell_path.is_empty() {
    ShellEnvironmentStatus::LoginPathUnavailable
  } else {
    ShellEnvironmentStatus::Ready
  };

  (login_shell_path, status)
}

#[cfg(target_os = "macos")]
async fn detect_default_shell(home_dir: std::path::PathBuf) -> Option<String> {
  let home = home_dir.to_string_lossy().to_string();
  if !home.starts_with('/') {
    return None;
  }

  let output = probe("/usr/bin/dscl", &[".", "-read", home.as_str(), "UserShell"]).await?;
  output.lines().find_map(|line| {
    let shell = line.strip_prefix("UserShell:")?.trim();
    (shell.starts_with('/') && shell.len() > 1).then(|| shell.to_string())
  })
}

#[cfg(target_os = "linux")]
async fn detect_default_shell(home_dir: std::path::PathBuf) -> Option<String> {
  tokio::task::spawn_blocking(move || {
    let contents = std::fs::read_to_string("/etc/passwd").ok()?;
    contents.lines().find_map(|line| {
      let mut fields = line.split(':');
      let (_name, _password, _uid, _gid, _gecos, home, shell) = (
        fields.next()?,
        fields.next()?,
        fields.next()?,
        fields.next()?,
        fields.next()?,
        fields.next()?,
        fields.next()?,
      );
      (home_dir.as_path() == Path::new(home) && shell.starts_with('/') && shell.len() > 1)
        .then(|| shell.to_string())
    })
  })
  .await
  .ok()
  .flatten()
}

#[cfg(not(any(target_os = "macos", target_os = "linux")))]
async fn detect_default_shell(_home_dir: std::path::PathBuf) -> Option<String> {
  None
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
