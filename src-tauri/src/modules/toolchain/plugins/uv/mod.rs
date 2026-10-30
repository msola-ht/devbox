use std::path::Path;

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor, ToolKind};
use super::super::plugin::ToolPlugin;

static UV_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "uv",
  display_name: "uv",
  description: "跨平台 Python 包与项目管理工具",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
  kind: ToolKind::Tool,
};

pub struct UvPlugin;

impl ToolPlugin for UvPlugin {
  fn descriptor(&self) -> &'static ToolDescriptor {
    &UV_DESCRIPTOR
  }

  fn executable_names(&self) -> &'static [&'static str] {
    #[cfg(windows)]
    {
      &["uv.exe"]
    }
    #[cfg(not(windows))]
    {
      &["uv"]
    }
  }

  fn version_args(&self) -> &'static [&'static str] {
    &["--version"]
  }

  fn parse_version_output(&self, output: &str) -> Option<String> {
    output.lines().find_map(|line| {
      let mut tokens = line.split_whitespace();
      if tokens.next()? != "uv" {
        return None;
      }
      tokens
        .next()
        .filter(|version| version.chars().next().is_some_and(|ch| ch.is_ascii_digit()))
        .map(str::to_string)
    })
  }

  fn source_hint(&self, path: &Path) -> Option<String> {
    let path = path.to_string_lossy();
    if path.starts_with("/opt/homebrew/") || path.starts_with("/usr/local/Cellar/") {
      Some("Homebrew".to_string())
    } else if path.starts_with("/usr/local/") {
      Some("系统路径".to_string())
    } else {
      Some("PATH".to_string())
    }
  }
}
