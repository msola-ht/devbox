use std::path::{Path, PathBuf};

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor};
use super::super::plugin::{ToolPlugin, ToolPluginContext};

static NODE_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "node",
  display_name: "Node.js",
  description: "JavaScript 运行时与包管理基础环境",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
};

pub struct NodePlugin;

impl ToolPlugin for NodePlugin {
  fn descriptor(&self) -> &'static ToolDescriptor {
    &NODE_DESCRIPTOR
  }

  fn executable_names(&self) -> &'static [&'static str] {
    #[cfg(windows)]
    {
      &["node.exe"]
    }
    #[cfg(not(windows))]
    {
      &["node"]
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
    Some(version_manager_source(path).unwrap_or_else(|| {
      let path = normalized_path(path);
      if path.contains("/homebrew/") || path.contains("/cellar/") {
        "Homebrew".to_string()
      } else if path.starts_with("/usr/") || path.contains("/program files/nodejs/") {
        "系统安装".to_string()
      } else {
        "PATH".to_string()
      }
    }))
  }
}

pub(super) fn node_version_manager_directories(
  environment: &ToolPluginContext<'_>,
) -> Vec<PathBuf> {
  let home = &environment.home_dir;
  let directories = vec![
    home.join(".volta/bin"),
    home.join(".nodenv/shims"),
    home.join(".asdf/shims"),
    home.join(".local/share/fnm/aliases/default/bin"),
    home.join(".fnm/aliases/default/bin"),
    home.join(".nvm/current/bin"),
  ];

  #[cfg(windows)]
  let directories = {
    let mut directories = directories;
    if let Some(local_app_data) = std::env::var_os("LOCALAPPDATA") {
      let local_app_data = PathBuf::from(local_app_data);
      directories.push(local_app_data.join("fnm/aliases/default/bin"));
      directories.push(local_app_data.join("Volta/bin"));
    }
    directories
  };

  directories
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

fn version_manager_source(path: &Path) -> Option<String> {
  let path = normalized_path(path);
  if path.contains("/fnm/") {
    Some("fnm".to_string())
  } else if path.contains("/.nvm/") {
    Some("nvm".to_string())
  } else if path.contains("/.volta/") {
    Some("Volta".to_string())
  } else if path.contains("/.nodenv/") {
    Some("nodenv".to_string())
  } else if path.contains("/.asdf/") {
    Some("asdf".to_string())
  } else {
    None
  }
}

fn normalized_path(path: &Path) -> String {
  path.to_string_lossy().replace('\\', "/").to_lowercase()
}
