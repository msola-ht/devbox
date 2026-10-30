use std::path::Path;

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor, ToolKind};
use super::super::plugin::ToolPlugin;
use super::super::plugin_support::{normalized_path, parse_version_line};

static NODE_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "node",
  display_name: "Node.js",
  description: "JavaScript 运行时与包管理基础环境",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
  kind: ToolKind::Tool,
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
