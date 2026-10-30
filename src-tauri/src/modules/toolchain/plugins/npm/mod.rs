use std::path::{Path, PathBuf};

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor};
use super::super::plugin::{ToolPlugin, ToolPluginContext};
use super::node::node_version_manager_directories;

static NPM_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "npm",
  display_name: "npm",
  description: "Node.js 官方包管理器",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
};

pub struct NpmPlugin;

impl ToolPlugin for NpmPlugin {
  fn descriptor(&self) -> &'static ToolDescriptor {
    &NPM_DESCRIPTOR
  }

  fn executable_names(&self) -> &'static [&'static str] {
    #[cfg(windows)]
    {
      &["npm.cmd", "npm"]
    }
    #[cfg(not(windows))]
    {
      &["npm"]
    }
  }

  fn known_directories(&self, environment: &ToolPluginContext<'_>) -> Vec<PathBuf> {
    let mut directories = match environment.os {
      "macos" => vec![
        PathBuf::from("/opt/homebrew/bin"),
        PathBuf::from("/usr/local/bin"),
      ],
      "linux" => vec![
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/usr/bin"),
        PathBuf::from("/bin"),
      ],
      "windows" => std::env::var_os("ProgramFiles")
        .map(|path| vec![PathBuf::from(path).join("nodejs")])
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
    if normalized.contains("/fnm/")
      || normalized.contains("/.nvm/")
      || normalized.contains("/nodejs/")
    {
      Some("Node.js 自带".to_string())
    } else if normalized.contains("/npm-global/")
      || normalized.contains("/.npm-global/")
    {
      Some("npm 全局安装".to_string())
    } else if normalized.contains("/homebrew/") || normalized.contains("/cellar/") {
      Some("Homebrew".to_string())
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
