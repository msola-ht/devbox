use std::path::Path;

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor, ToolKind};
use super::super::plugin::ToolPlugin;
use super::super::plugin_support::{normalized_path, parse_version_token};

static GIT_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "git",
  display_name: "Git",
  description: "分布式版本控制工具",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
  kind: ToolKind::Tool,
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
