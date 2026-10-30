use std::path::{Path, PathBuf};

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor};
use super::super::plugin::{ToolPlugin, ToolPluginContext};
use super::node::node_version_manager_directories;

static PNPM_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "pnpm",
  display_name: "pnpm",
  description: "节省磁盘空间的 Node.js 包管理器",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
};

pub struct PnpmPlugin;

impl ToolPlugin for PnpmPlugin {
  fn descriptor(&self) -> &'static ToolDescriptor {
    &PNPM_DESCRIPTOR
  }

  fn executable_names(&self) -> &'static [&'static str] {
    #[cfg(windows)]
    {
      &["pnpm.cmd", "pnpm"]
    }
    #[cfg(not(windows))]
    {
      &["pnpm"]
    }
  }

  fn known_directories(&self, environment: &ToolPluginContext<'_>) -> Vec<PathBuf> {
    let home = &environment.home_dir;
    let mut directories = match environment.os {
      "macos" => vec![
        PathBuf::from("/opt/homebrew/bin"),
        PathBuf::from("/usr/local/bin"),
        home.join("Library/pnpm"),
        home.join(".local/share/pnpm"),
        home.join(".local/share/npm-global/bin"),
        home.join(".pnpm"),
      ],
      "linux" => vec![
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/usr/bin"),
        home.join(".local/share/pnpm"),
        home.join(".local/share/npm-global/bin"),
        home.join(".local/bin"),
      ],
      "windows" => std::env::var_os("LOCALAPPDATA")
        .map(|path| vec![PathBuf::from(path).join("pnpm")])
        .unwrap_or_default(),
      _ => Vec::new(),
    };

    directories.extend(node_version_manager_directories(environment));
    directories
  }

  fn version_args(&self) -> &'static [&'static str] {
    &["--version"]
  }

  fn parse_version_output(&self, output: &str) -> Option<String> {
    parse_version_line(output)
  }

  fn source_hint(&self, path: &Path) -> Option<String> {
    let normalized = normalized_path(path);
    if normalized.contains("/npm-global/") || normalized.contains("/.npm-global/") {
      Some("npm 全局安装".to_string())
    } else if normalized.contains("/homebrew/") || normalized.contains("/cellar/") {
      Some("Homebrew".to_string())
    } else if normalized.contains("/pnpm/") {
      Some("pnpm 独立安装".to_string())
    } else {
      Some("PATH".to_string())
    }
  }
}

fn parse_version_line(output: &str) -> Option<String> {
  output.lines().find_map(|line| {
    line.split_whitespace()
      .find_map(parse_version_token)
  })
}

fn parse_version_token(token: &str) -> Option<String> {
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
}

fn normalized_path(path: &Path) -> String {
  path.to_string_lossy().replace('\\', "/").to_lowercase()
}
