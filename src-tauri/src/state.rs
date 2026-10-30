use std::sync::Arc;

use uuid::Uuid;

use crate::modules::environment::EnvironmentService;
use crate::modules::settings::SettingsService;
use crate::modules::toolchain::ToolchainService;

#[derive(Clone)]
pub struct AppState {
  pub boot_id: String,
  pub environment: Arc<EnvironmentService>,
  pub settings: Arc<SettingsService>,
  pub toolchain: Arc<ToolchainService>,
}

impl AppState {
  pub fn new(
    environment: EnvironmentService,
    settings: SettingsService,
    toolchain: ToolchainService,
  ) -> Self {
    Self {
      boot_id: Uuid::new_v4().to_string(),
      environment: Arc::new(environment),
      settings: Arc::new(settings),
      toolchain: Arc::new(toolchain),
    }
  }
}
