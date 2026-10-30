use std::fmt;
use std::path::Path;
use std::process::Stdio;
use std::time::Duration;

use tokio::io::{AsyncRead, AsyncReadExt};
use tokio::process::Command;
use tokio::time::timeout;

const MAX_OUTPUT_BYTES: usize = 64 * 1024;

#[derive(Debug)]
pub enum VersionProbeError {
  Spawn(String),
  Io(String),
  Timeout,
  OutputTooLarge,
  ExitFailure(String),
}

impl fmt::Display for VersionProbeError {
  fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
    match self {
      Self::Spawn(message) => write!(formatter, "无法启动版本命令: {message}"),
      Self::Io(message) => write!(formatter, "版本命令 I/O 失败: {message}"),
      Self::Timeout => formatter.write_str("版本命令执行超时"),
      Self::OutputTooLarge => formatter.write_str("版本命令输出超过限制"),
      Self::ExitFailure(message) => write!(formatter, "版本命令执行失败: {message}"),
    }
  }
}

pub async fn probe_version(
  executable: &Path,
  args: &[&str],
  command_timeout: Duration,
) -> Result<String, VersionProbeError> {
  let mut command = Command::new(executable);
  command
    .args(args)
    .stdin(Stdio::null())
    .stdout(Stdio::piped())
    .stderr(Stdio::piped())
    .kill_on_drop(true);
  configure_environment(&mut command);

  let mut child = command
    .spawn()
    .map_err(|error| VersionProbeError::Spawn(error.to_string()))?;
  let stdout = child
    .stdout
    .take()
    .ok_or_else(|| VersionProbeError::Io("无法读取标准输出".to_string()))?;
  let stderr = child
    .stderr
    .take()
    .ok_or_else(|| VersionProbeError::Io("无法读取标准错误".to_string()))?;

  let stdout_task = tokio::spawn(read_limited(stdout));
  let stderr_task = tokio::spawn(read_limited(stderr));

  let status = match timeout(command_timeout, child.wait()).await {
    Ok(Ok(status)) => status,
    Ok(Err(error)) => {
      stdout_task.abort();
      stderr_task.abort();
      return Err(VersionProbeError::Io(error.to_string()));
    }
    Err(_) => {
      let _ = child.kill().await;
      let _ = child.wait().await;
      stdout_task.abort();
      stderr_task.abort();
      return Err(VersionProbeError::Timeout);
    }
  };

  let (stdout, stdout_too_large) = stdout_task
    .await
    .map_err(|error| VersionProbeError::Io(error.to_string()))?
    .map_err(|error| VersionProbeError::Io(error.to_string()))?;
  let (stderr, stderr_too_large) = stderr_task
    .await
    .map_err(|error| VersionProbeError::Io(error.to_string()))?
    .map_err(|error| VersionProbeError::Io(error.to_string()))?;

  if stdout_too_large || stderr_too_large {
    return Err(VersionProbeError::OutputTooLarge);
  }

  if !status.success() {
    let message = String::from_utf8_lossy(&stderr).trim().to_string();
    return Err(VersionProbeError::ExitFailure(if message.is_empty() {
      format!("退出码 {status}")
    } else {
      message
    }));
  }

  let output = if stdout.is_empty() { stderr } else { stdout };
  Ok(String::from_utf8_lossy(&output).trim().to_string())
}

fn configure_environment(command: &mut Command) {
  command.env_clear();

  #[cfg(windows)]
  for variable in [
    "SystemRoot",
    "WINDIR",
    "TEMP",
    "TMP",
    "USERPROFILE",
    "LOCALAPPDATA",
    "APPDATA",
    "PATHEXT",
    "COMSPEC",
  ] {
    if let Some(value) = std::env::var_os(variable) {
      command.env(variable, value);
    }
  }
}

async fn read_limited<R>(reader: R) -> std::io::Result<(Vec<u8>, bool)>
where
  R: AsyncRead + Unpin,
{
  let mut buffer = Vec::new();
  reader
    .take((MAX_OUTPUT_BYTES + 1) as u64)
    .read_to_end(&mut buffer)
    .await?;
  let too_large = buffer.len() > MAX_OUTPUT_BYTES;
  buffer.truncate(MAX_OUTPUT_BYTES);
  Ok((buffer, too_large))
}
