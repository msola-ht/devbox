pub const TOOL_PLUGIN_CONTRACT_VERSION: u32 = 1;

#[derive(Clone, Copy, Debug)]
pub enum InstallationSource {
  System,
  Managed,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ToolKind {
  Tool,
  Shell,
}

#[derive(Clone, Debug)]
pub struct ToolDescriptor {
  pub tool_id: &'static str,
  pub display_name: &'static str,
  pub description: &'static str,
  pub contract_version: u32,
  pub kind: ToolKind,
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
  pub kind: ToolKind,
  pub is_default_shell: bool,
  pub effective_installation: Option<InstallationSummary>,
  pub shell_installations: Vec<InstallationSummary>,
  pub managed_installations: Vec<InstallationSummary>,
  pub active_installation_id: Option<String>,
  pub active_installation_error_code: Option<String>,
}

#[derive(Clone, Debug)]
pub enum ActiveInstallationState {
  Absent,
  Selected(String),
  Invalid(&'static str),
}

impl ToolSummary {
  pub fn empty(descriptor: &ToolDescriptor) -> Self {
    Self {
      tool_id: descriptor.tool_id.to_string(),
      display_name: descriptor.display_name.to_string(),
      description: descriptor.description.to_string(),
      plugin_contract_version: descriptor.contract_version.to_string(),
      kind: descriptor.kind,
      is_default_shell: false,
      effective_installation: None,
      shell_installations: Vec::new(),
      managed_installations: Vec::new(),
      active_installation_id: None,
      active_installation_error_code: None,
    }
  }

  pub fn set_effective_installation(&mut self, installation: Option<InstallationSummary>) {
    self.effective_installation = installation;
  }

  pub fn set_shell_state(&mut self, installations: Vec<InstallationSummary>, is_default: bool) {
    self.is_default_shell = self.kind == ToolKind::Shell && is_default;
    self.shell_installations = installations;
  }

  pub fn set_managed_state(
    &mut self,
    installations: Vec<InstallationSummary>,
    active: ActiveInstallationState,
  ) {
    self.managed_installations = installations;
    self.active_installation_id = None;
    self.active_installation_error_code = None;

    match active {
      ActiveInstallationState::Absent => {}
      ActiveInstallationState::Selected(installation_id) => {
        let is_healthy = self.managed_installations.iter().any(|installation| {
          installation.installation_id == installation_id && installation.healthy
        });
        if is_healthy {
          self.active_installation_id = Some(installation_id);
        } else {
          self.active_installation_error_code = Some("ACTIVE_INSTALLATION_UNAVAILABLE".to_string());
        }
      }
      ActiveInstallationState::Invalid(code) => {
        self.active_installation_error_code = Some(code.to_string());
      }
    }
  }
}

#[derive(Clone, Debug)]
pub struct ToolchainSnapshot {
  pub revision: u64,
  pub environment_revision: u64,
  pub tools: Vec<ToolSummary>,
}
