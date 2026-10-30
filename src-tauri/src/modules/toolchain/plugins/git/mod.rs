use std::path::{Path, PathBuf};

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor};
use super::super::plugin::{ToolPlugin, ToolPluginContext};

static GIT_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "git",
  display_name: "Git",
  description: "分布式版本控制工具",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
};

pub struct GitPlugin;

impl ToolPlugin for GitPlugin {
  fn descriptor(&self) -> &'static ToolDescriptor {
    &GIT_DESCRIPTOR
  }

  fn executable_names(&self) -> &'static [&'static str] {
    #[cfg(windows)]
    {
      &["git.exe"]
    }
    #[cfg(not(windows))]
    {
      &["git"]
    }
  }

  fn known_directories(&self, environment: &ToolPluginContext<'_>) -> Vec<PathBuf> {
    let mut directories = match environment.os {
      "macos" => vec![
        PathBuf::from("/opt/homebrew/bin"),
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/usr/bin"),
      ],
      "linux" => vec![
        PathBuf::from("/usr/local/bin"),
        PathBuf::from("/usr/bin"),
        PathBuf::from("/bin"),
      ],
      "windows" => {
        let mut directories = std::env::var_os("ProgramFiles")
          .map(|path| vec![PathBuf::from(path).join("Git/cmd")])
          .unwrap_or_default();
        if let Some(local_app_data) = std::env::var_os("LOCALAPPDATA") {
          directories.push(PathBuf::from(local_app_data).join("Programs/Git/cmd"));
        }
        directories
      }
      _ => Vec::new(),
    };

    directories.push(environment.home_dir.join("scoop/shims"));
    directories
  }

  fn version_args(&self) -> &'static [&'static str] {
    &["--version"]
  }

  fn parse_version_output(&self, output: &str) -> Option<String> {
    output.lines().find_map(|line| {
      let mut tokens = line.split_whitespace();
      if tokens.next()? != "git" || tokens.next()? != "version" {
        return None;
      }
      tokens.next().and_then(parse_version_token)
    })
  }

  fn source_hint(&self, path: &Path) -> Option<String> {
    let normalized = normalized_path(path);
    if normalized.contains("/homebrew/") || normalized.contains("/cellar/") {
      Some("Homebrew".to_string())
    } else if (cfg!(target_os = "macos") && normalized.starts_with("/usr/bin/git"))
      || normalized.contains("/applications/xcode")
    {
      Some("Apple Git".to_string())
    } else if normalized.contains("/scoop/") {
      Some("Scoop".to_string())
    } else if normalized.contains("/program files/git/") {
      Some("Git for Windows".to_string())
    } else if normalized.starts_with("/usr/bin/") || normalized.starts_with("/bin/") {
      Some("系统安装".to_string())
    } else {
      Some("PATH".to_string())
    }
  }
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
