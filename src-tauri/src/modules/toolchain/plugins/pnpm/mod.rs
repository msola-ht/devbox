use std::path::Path;

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor, ToolKind};
use super::super::plugin::ToolPlugin;
use super::super::plugin_support::{normalized_path, parse_version_line};

static PNPM_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "pnpm",
  display_name: "pnpm",
  description: "节省磁盘空间的 Node.js 包管理器",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
  kind: ToolKind::Tool,
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
