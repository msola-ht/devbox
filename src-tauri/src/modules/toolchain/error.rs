use std::error::Error;
use std::fmt;
use std::io;
use std::path::PathBuf;

#[derive(Debug)]
pub enum ToolchainError {
  UnsafeManagedToolPath {
    path: PathBuf,
  },
  ReadManagedToolRoot {
    path: PathBuf,
    source: io::Error,
  },
  ReadManagedToolEntries {
    path: PathBuf,
    source: io::Error,
  },
  ReadManagedToolEntry {
    source: io::Error,
  },
  ReadManagedToolEntryType {
    source: io::Error,
  },
  CanonicalizeManagedToolRoot {
    path: PathBuf,
    source: io::Error,
  },
  BlockingTaskFailed {
    operation: &'static str,
    message: String,
  },
}

impl fmt::Display for ToolchainError {
  fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
    match self {
      Self::UnsafeManagedToolPath { path } => {
        write!(formatter, "托管工具路径不是安全目录: {}", path.display())
      }
      Self::ReadManagedToolRoot { path, source } => write!(
        formatter,
        "无法读取托管工具目录 {}: {source}",
        path.display()
      ),
      Self::ReadManagedToolEntries { path, source } => write!(
        formatter,
        "无法读取托管工具目录 {}: {source}",
        path.display()
      ),
      Self::ReadManagedToolEntry { source } => {
        write!(formatter, "无法读取托管安装条目: {source}")
      }
      Self::ReadManagedToolEntryType { source } => {
        write!(formatter, "无法读取托管安装条目类型: {source}")
      }
      Self::CanonicalizeManagedToolRoot { path, source } => write!(
        formatter,
        "无法规范化托管工具目录 {}: {source}",
        path.display()
      ),
      Self::BlockingTaskFailed { operation, message } => {
        write!(formatter, "{operation}阻塞任务失败: {message}")
      }
    }
  }
}

impl Error for ToolchainError {
  fn source(&self) -> Option<&(dyn Error + 'static)> {
    match self {
      Self::ReadManagedToolRoot { source, .. }
      | Self::ReadManagedToolEntries { source, .. }
      | Self::ReadManagedToolEntry { source }
      | Self::ReadManagedToolEntryType { source }
      | Self::CanonicalizeManagedToolRoot { source, .. } => Some(source),
      Self::UnsafeManagedToolPath { .. } | Self::BlockingTaskFailed { .. } => None,
    }
  }
}
