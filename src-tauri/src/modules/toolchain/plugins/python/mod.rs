use std::path::{Path, PathBuf};

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor};
use super::super::plugin::{ToolPlugin, ToolPluginContext};

static PYTHON_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "python",
  display_name: "Python",
  description: "Python 解释器与运行时",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
};

pub struct PythonPlugin;

impl ToolPlugin for PythonPlugin {
  fn descriptor(&self) -> &'static ToolDescriptor {
    &PYTHON_DESCRIPTOR
  }

  fn executable_names(&self) -> &'static [&'static str] {
    #[cfg(windows)]
    {
      &["python.exe", "python3.exe"]
    }
    #[cfg(not(windows))]
    {
      &["python3", "python"]
    }
  }

  fn known_directories(&self, environment: &ToolPluginContext<'_>) -> Vec<PathBuf> {
    match environment.os {
      "macos" => vec![
        PathBuf::from("/opt/homebrew/bin"),
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/Library/Frameworks/Python.framework/Versions/Current/bin"),
        environment.home_dir.join(".local/bin"),
        environment.home_dir.join(".pyenv/shims"),
      ],
      "linux" => vec![
        environment.home_dir.join(".local/bin"),
        environment.home_dir.join(".pyenv/shims"),
        PathBuf::from("/home/linuxbrew/.linuxbrew/bin"),
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/usr/bin"),
        PathBuf::from("/bin"),
      ],
      "windows" => vec![environment.home_dir.join(".local").join("bin")],
      _ => Vec::new(),
    }
  }

  fn version_args(&self) -> &'static [&'static str] {
    &["--version"]
  }

  fn parse_version_output(&self, output: &str) -> Option<String> {
    output.lines().find_map(|line| {
      let mut tokens = line.split_whitespace();
      if tokens.next()? != "Python" {
        return None;
      }

      tokens
        .next()
        .filter(|version| {
          version.contains('.')
            && version
              .chars()
              .all(|character| character.is_ascii_digit() || character == '.')
            && version
              .chars()
              .next()
              .is_some_and(|character| character.is_ascii_digit())
        })
        .map(str::to_string)
    })
  }

  fn source_hint(&self, path: &Path) -> Option<String> {
    let path = path.to_string_lossy().replace('\\', "/");
    if path.contains("/.pyenv/") {
      Some("pyenv".to_string())
    } else if path.contains("/.local/share/uv/python/") {
      Some("uv 管理".to_string())
    } else if path.contains("/opt/homebrew/")
      || path.contains("/usr/local/Cellar/")
      || path.contains("/home/linuxbrew/")
    {
      Some("Homebrew".to_string())
    } else if path.starts_with("/usr/bin/") || path.starts_with("/bin/") {
      Some("系统 Python".to_string())
    } else if path.contains("/.venv/") || path.contains("/venv/") {
      Some("虚拟环境".to_string())
    } else if path.contains("/.local/") {
      Some("用户安装".to_string())
    } else {
      Some("PATH".to_string())
    }
  }
}
