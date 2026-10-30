use std::path::{Path, PathBuf};

use super::model::ToolDescriptor;

pub struct ToolPluginContext<'a> {
  pub os: &'a str,
  pub home_dir: &'a Path,
  pub login_shell_path: &'a [PathBuf],
  pub default_shell: Option<&'a str>,
}

pub trait ToolPlugin: Send + Sync {
  fn descriptor(&self) -> &'static ToolDescriptor;
  fn executable_names(&self) -> &'static [&'static str];
  fn preferred_directories(&self, _environment: &ToolPluginContext<'_>) -> Vec<PathBuf> {
    Vec::new()
  }
  fn known_directories(&self, environment: &ToolPluginContext<'_>) -> Vec<PathBuf>;
  fn version_args(&self) -> &'static [&'static str];
  fn parse_version_output(&self, output: &str) -> Option<String>;
  fn source_hint(&self, path: &Path) -> Option<String>;
}
