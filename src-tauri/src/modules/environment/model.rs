use std::path::PathBuf;

#[derive(Clone, Debug)]
pub struct EnvironmentInfo {
  pub os: String,
  pub arch: String,
  pub home_dir: PathBuf,
  pub devbox_root: PathBuf,
  pub devbox_ready: bool,
  pub devbox_error: Option<String>,
}
