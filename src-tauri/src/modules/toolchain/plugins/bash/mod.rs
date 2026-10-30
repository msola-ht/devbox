use std::path::{Path, PathBuf};

use super::super::model::{TOOL_PLUGIN_CONTRACT_VERSION, ToolDescriptor, ToolKind};
use super::super::plugin::{ToolPlugin, ToolPluginContext};
use super::super::plugin_support::{default_shell_candidate, normalized_path};

static BASH_DESCRIPTOR: ToolDescriptor = ToolDescriptor {
  tool_id: "bash",
  display_name: "Bash",
  description: "GNU Bourne-Again Shell",
  contract_version: TOOL_PLUGIN_CONTRACT_VERSION,
  kind: ToolKind::Shell,
};

pub struct BashPlugin;

impl ToolPlugin for BashPlugin {
  fn descriptor(&self) -> &'static ToolDescriptor {
    &BASH_DESCRIPTOR
  }

  fn executable_names(&self) -> &'static [&'static str] {
    #[cfg(windows)]
    {
      &["bash.exe"]
    }
    #[cfg(not(windows))]
    {
      &["bash"]
    }
  }

  fn current_candidate_override(&self, context: &ToolPluginContext<'_>) -> Option<PathBuf> {
    default_shell_candidate("bash", context.default_shell)
  }

  fn version_args(&self) -> &'static [&'static str] {
    &["--version"]
  }

  fn parse_version_output(&self, output: &str) -> Option<String> {
    let line = output.lines().find(|line| line.contains("version"))?;
    let raw_version = line.split_once("version")?.1.trim();
    let mut characters = raw_version.chars();
    if !characters.next()?.is_ascii_digit() {
      return None;
    }

    let version = raw_version
      .chars()
      .take_while(|character| {
        character.is_ascii_alphanumeric() || matches!(character, '.' | '-' | '+')
      })
      .collect::<String>();
    version.contains('.').then_some(version)
  }

  fn source_hint(&self, path: &Path) -> Option<String> {
    let normalized = normalized_path(path);
    if normalized.contains("/homebrew/") || normalized.contains("/cellar/") {
      Some("Homebrew".to_string())
    } else if normalized.contains("/git/bin/") {
      Some("Git for Windows".to_string())
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
