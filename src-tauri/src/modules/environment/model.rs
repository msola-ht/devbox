use std::path::PathBuf;

#[derive(Clone, Debug)]
pub struct EnvironmentInfo {
  pub os: String,
  pub os_version: Option<String>,
  pub os_build: Option<String>,
  pub arch: String,
  pub memory_total_bytes: Option<u64>,
  pub disk_total_bytes: Option<u64>,
  pub gpu_name: Option<String>,
  pub gpu_core_count: Option<u32>,
  pub default_shell: Option<String>,
  pub login_shell_path: Vec<PathBuf>,
  pub home_dir: PathBuf,
  pub devbox_root: PathBuf,
  pub devbox_ready: bool,
  pub devbox_error: Option<String>,
}
