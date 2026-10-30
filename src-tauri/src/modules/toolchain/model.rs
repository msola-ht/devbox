pub const TOOL_PLUGIN_CONTRACT_VERSION: u32 = 1;

#[derive(Clone, Copy, Debug)]
pub enum InstallationSource {
  System,
  Managed,
}

#[derive(Clone, Debug)]
pub struct ToolDescriptor {
  pub tool_id: &'static str,
  pub display_name: &'static str,
  pub description: &'static str,
  pub contract_version: u32,
}

#[derive(Clone, Debug)]
pub struct InstallationSummary {
  pub installation_id: String,
  pub version: Option<String>,
  pub path: String,
  pub source: InstallationSource,
  pub source_hint: Option<String>,
  pub healthy: bool,
  pub error_code: Option<String>,
}

#[derive(Clone, Debug)]
pub struct ToolSummary {
  pub tool_id: String,
  pub display_name: String,
  pub description: String,
  pub plugin_contract_version: String,
  pub system_installations: Vec<InstallationSummary>,
  pub managed_installations: Vec<InstallationSummary>,
  pub active_installation_id: Option<String>,
}

impl ToolSummary {
  pub fn empty(descriptor: &ToolDescriptor) -> Self {
    Self {
      tool_id: descriptor.tool_id.to_string(),
      display_name: descriptor.display_name.to_string(),
      description: descriptor.description.to_string(),
      plugin_contract_version: descriptor.contract_version.to_string(),
      system_installations: Vec::new(),
      managed_installations: Vec::new(),
      active_installation_id: None,
    }
  }
}

#[derive(Clone, Debug)]
pub struct ToolchainSnapshot {
  pub revision: u64,
  pub tools: Vec<ToolSummary>,
}
