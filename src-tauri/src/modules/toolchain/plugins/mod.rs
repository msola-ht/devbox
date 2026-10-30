mod python;
mod uv;

use super::plugin::ToolPlugin;
pub use python::PythonPlugin;
pub use uv::UvPlugin;

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
