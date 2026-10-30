use std::sync::RwLock;

use super::model::{EnvironmentInfo, EnvironmentSnapshot};
use super::system::detect_system_profile;

pub struct EnvironmentService {
  snapshot: RwLock<EnvironmentSnapshot>,
  refresh_lock: tokio::sync::Mutex<()>,
}

impl EnvironmentService {
  pub fn new(info: EnvironmentInfo) -> Self {
    Self {
      snapshot: RwLock::new(EnvironmentSnapshot { revision: 1, info }),
      refresh_lock: tokio::sync::Mutex::new(()),
    }
  }

  pub fn snapshot(&self) -> EnvironmentSnapshot {
    match self.snapshot.read() {
      Ok(snapshot) => snapshot.clone(),
      Err(poisoned) => poisoned.into_inner().clone(),
    }
  }

  pub async fn refresh(&self) -> EnvironmentSnapshot {
    let _refresh_guard = self.refresh_lock.lock().await;
    let previous = self.snapshot();
    let profile = detect_system_profile(&previous.info.home_dir).await;
    let next = EnvironmentInfo {
      os: previous.info.os,
      os_version: profile.os_version,
      os_build: profile.os_build,
      arch: previous.info.arch,
      memory_total_bytes: profile.memory_total_bytes,
      disk_total_bytes: profile.disk_total_bytes,
      gpu_name: profile.gpu_name,
      gpu_core_count: profile.gpu_core_count,
      default_shell: profile.default_shell,
      shell_environment_status: profile.shell_environment_status,
      login_shell_path: profile.login_shell_path,
      home_dir: previous.info.home_dir,
      devbox_root: previous.info.devbox_root,
      devbox_ready: previous.info.devbox_ready,
      devbox_error: previous.info.devbox_error,
    };
    let next = EnvironmentSnapshot {
      revision: previous.revision + 1,
      info: next,
    };

    match self.snapshot.write() {
      Ok(mut snapshot) => *snapshot = next.clone(),
      Err(poisoned) => *poisoned.into_inner() = next.clone(),
    }

    log::info!(
      "环境刷新完成 revision={} default_shell={:?} shell_environment_status={:?} login_shell_path_entries={}",
      next.revision,
      next.info.default_shell,
      next.info.shell_environment_status,
      next.info.login_shell_path.len()
    );

    next
  }
}
