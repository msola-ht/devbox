use std::path::{Path, PathBuf};

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor};
use super::super::plugin::{ToolPlugin, ToolPluginContext};

static ZSH_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "zsh",
  display_name: "Zsh",
  description: "Z Shell 交互式命令行",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
};

pub struct ZshPlugin;

impl ToolPlugin for ZshPlugin {
  fn descriptor(&self) -> &'static ToolDescriptor {
    &ZSH_DESCRIPTOR
  }

  fn executable_names(&self) -> &'static [&'static str] {
    #[cfg(windows)]
    {
      &["zsh.exe"]
    }
    #[cfg(not(windows))]
    {
      &["zsh"]
    }
  }

  fn known_directories(&self, environment: &ToolPluginContext<'_>) -> Vec<PathBuf> {
    match environment.os {
      "macos" => vec![
        PathBuf::from("/opt/homebrew/bin"),
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/bin"),
      ],
      "linux" => vec![
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/usr/bin"),
        PathBuf::from("/bin"),
      ],
      _ => Vec::new(),
    }
  }

  fn version_args(&self) -> &'static [&'static str] {
    &["--version"]
  }

  fn parse_version_output(&self, output: &str) -> Option<String> {
    output.lines().find_map(|line| {
      let mut tokens = line.split_whitespace();
      if tokens.next()? != "zsh" {
        return None;
      }
      let token = tokens.next()?;
      let version = token.strip_prefix('v').unwrap_or(token);
      let mut characters = version.chars();
      if !characters.next()?.is_ascii_digit()
        || !version.contains('.')
        || !characters.all(|character| {
          character.is_ascii_alphanumeric()
            || matches!(character, '.' | '-' | '+')
        })
      {
        return None;
      }
      Some(version.to_string())
    })
  }

  fn source_hint(&self, path: &Path) -> Option<String> {
    let normalized = normalized_path(path);
    if normalized.contains("/homebrew/") || normalized.contains("/cellar/") {
      Some("Homebrew".to_string())
    } else if normalized.starts_with("/bin/")
      || normalized.starts_with("/usr/bin/")
      || normalized.starts_with("/usr/local/bin/")
    {
      Some("系统 Shell".to_string())
    } else {
      Some("PATH".to_string())
    }
  }
}

fn normalized_path(path: &Path) -> String {
  path.to_string_lossy().replace('\\', "/").to_lowercase()
}
