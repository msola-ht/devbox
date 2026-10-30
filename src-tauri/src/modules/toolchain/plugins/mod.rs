mod bash;
mod git;
mod node;
mod npm;
mod pnpm;
mod python;
mod uv;
mod zsh;

use super::plugin::ToolPlugin;
pub use bash::BashPlugin;
pub use git::GitPlugin;
pub use node::NodePlugin;
pub use npm::NpmPlugin;
pub use pnpm::PnpmPlugin;
pub use python::PythonPlugin;
pub use uv::UvPlugin;
pub use zsh::ZshPlugin;

pub struct PluginRegistry {
  plugins: Vec<Box<dyn ToolPlugin>>,
}

impl PluginRegistry {
  pub fn new(plugins: Vec<Box<dyn ToolPlugin>>) -> Self {
    Self { plugins }
  }

  pub fn plugins(&self) -> impl Iterator<Item = &dyn ToolPlugin> + '_ {
    self.plugins.iter().map(|plugin| plugin.as_ref())
  }
}
