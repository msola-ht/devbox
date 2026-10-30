use std::path::Path;

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor, ToolKind};
use super::super::plugin::ToolPlugin;
use super::super::plugin_support::{normalized_path, parse_version_line};

static NPM_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "npm",
  display_name: "npm",
  description: "Node.js 官方包管理器",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
  kind: ToolKind::Tool,
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
    } else if normalized.contains("/npm-global/") || normalized.contains("/.npm-global/") {
      Some("npm 全局安装".to_string())
    } else if normalized.contains("/homebrew/") || normalized.contains("/cellar/") {
      Some("Homebrew".to_string())
    } else {
      Some("PATH".to_string())
    }
  }
}
