use std::path::{Path, PathBuf};

use super::model::ToolDescriptor;

pub struct ToolPluginContext<'a> {
  pub default_shell: Option<&'a str>,
}

pub trait ToolPlugin: Send + Sync {
  fn descriptor(&self) -> &'static ToolDescriptor;
  fn executable_names(&self) -> &'static [&'static str];
  fn current_candidate_override(&self, _context: &ToolPluginContext<'_>) -> Option<PathBuf> {
    None
  }
  fn version_args(&self) -> &'static [&'static str];
  fn parse_version_output(&self, output: &str) -> Option<String>;
  fn source_hint(&self, path: &Path) -> Option<String>;
}
