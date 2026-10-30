use std::sync::Arc;

use uuid::Uuid;

use crate::modules::environment::EnvironmentService;
use crate::modules::toolchain::ToolchainService;

#[derive(Clone)]
pub struct AppState {
  pub boot_id: String,
  pub environment: Arc<EnvironmentService>,
  pub toolchain: Arc<ToolchainService>,
}

impl AppState {
  pub fn new(environment: EnvironmentService, toolchain: ToolchainService) -> Self {
    Self {
      boot_id: Uuid::new_v4().to_string(),
      environment: Arc::new(environment),
      toolchain: Arc::new(toolchain),
    }
  }
}
