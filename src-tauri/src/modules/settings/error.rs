use std::error::Error;
use std::fmt;
use std::io;
use std::path::PathBuf;

#[derive(Debug)]
pub enum SettingsError {
  InvalidHost,
  InvalidPort,
  InvalidBypass,
  Conflict {
    expected: u64,
    actual: u64,
  },
  CorruptConfig {
    path: PathBuf,
  },
  UnsafeConfigPath {
    path: PathBuf,
  },
  UnsupportedVersion {
    version: u32,
  },
  ReadConfig {
    path: PathBuf,
    source: io::Error,
  },
  WriteConfig {
    path: PathBuf,
    source: io::Error,
  },
  BlockingTaskFailed {
    message: String,
  },
  Unavailable {
    code: &'static str,
    message: &'static str,
  },
}

impl SettingsError {
  pub fn code(&self) -> &'static str {
    match self {
      Self::InvalidHost | Self::InvalidPort | Self::InvalidBypass => "INVALID_INPUT",
      Self::Conflict { .. } => "SETTINGS_CONFLICT",
      Self::CorruptConfig { .. } => "SETTINGS_CORRUPT",
      Self::UnsafeConfigPath { .. } => "SETTINGS_PATH_UNSAFE",
      Self::UnsupportedVersion { .. } => "SETTINGS_VERSION_UNSUPPORTED",
      Self::ReadConfig { .. } => "SETTINGS_READ_FAILED",
      Self::WriteConfig { .. } => "SETTINGS_WRITE_FAILED",
      Self::BlockingTaskFailed { .. } => "SETTINGS_UNAVAILABLE",
      Self::Unavailable { code, .. } => code,
    }
  }

  pub fn user_message(&self) -> &'static str {
    match self {
      Self::InvalidHost => "代理主机格式无效",
      Self::InvalidPort => "代理端口必须在 1 到 65535 之间",
      Self::InvalidBypass => "代理绕过列表格式无效，支持域名、IP、CIDR 或 *",
      Self::Conflict { .. } => "设置已被其他窗口修改，请刷新后重试",
      Self::CorruptConfig { .. } => "DevBox 设置文件损坏，且备份无法恢复",
      Self::UnsafeConfigPath { .. } => "DevBox 设置路径不安全，已停止写入",
      Self::UnsupportedVersion { .. } => "设置文件版本高于当前应用支持范围",
      Self::ReadConfig { .. } => "无法读取 DevBox 设置文件",
      Self::WriteConfig { .. } => "无法保存 DevBox 设置",
      Self::BlockingTaskFailed { .. } => "无法完成 DevBox 设置操作",
      Self::Unavailable { message, .. } => message,
    }
  }
}

impl fmt::Display for SettingsError {
  fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
    match self {
      Self::InvalidHost => formatter.write_str("代理主机格式无效"),
      Self::InvalidPort => formatter.write_str("代理端口无效"),
      Self::InvalidBypass => formatter.write_str("代理绕过列表无效"),
      Self::Conflict { expected, actual } => write!(
        formatter,
        "设置版本冲突: expected={expected}, actual={actual}"
      ),
      Self::CorruptConfig { path } => {
        write!(formatter, "设置文件损坏: {}", path.display())
      }
      Self::UnsafeConfigPath { path } => {
        write!(formatter, "设置路径不安全: {}", path.display())
      }
      Self::UnsupportedVersion { version } => {
        write!(formatter, "不支持的设置格式版本: {version}")
      }
      Self::ReadConfig { path, source } => {
        write!(formatter, "无法读取设置文件 {}: {source}", path.display())
      }
      Self::WriteConfig { path, source } => {
        write!(formatter, "无法写入设置文件 {}: {source}", path.display())
      }
      Self::BlockingTaskFailed { message } => {
        write!(formatter, "设置阻塞任务失败: {message}")
      }
      Self::Unavailable { code, message } => {
        write!(formatter, "设置不可用({code}): {message}")
      }
    }
  }
}

impl Error for SettingsError {
  fn source(&self) -> Option<&(dyn Error + 'static)> {
    match self {
      Self::ReadConfig { source, .. } | Self::WriteConfig { source, .. } => Some(source),
      _ => None,
    }
  }
}
