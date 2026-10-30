mod model;
mod service;
mod system;

pub use model::{EnvironmentInfo, EnvironmentSnapshot, ShellEnvironmentStatus};
pub use service::EnvironmentService;
pub(crate) use system::detect_system_profile;
