use serde::Serialize;
use tauri::State;

use super::error::IpcError;
use super::now_rfc3339;
use crate::state::AppState;

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct EnvironmentSnapshot {
  boot_id: String,
  revision: String,
  generated_at: String,
  os: String,
  os_version: Option<String>,
  os_build: Option<String>,
  arch: String,
  home_dir: String,
  devbox_root: String,
  devbox_ready: bool,
  devbox_error: Option<String>,
  memory_total_bytes: Option<u64>,
  disk_total_bytes: Option<u64>,
  gpu_name: Option<String>,
  gpu_core_count: Option<u32>,
}

#[tauri::command]
pub fn get_environment(state: State<'_, AppState>) -> Result<EnvironmentSnapshot, IpcError> {
  let info = state.environment.info();

  Ok(EnvironmentSnapshot {
    boot_id: state.boot_id.clone(),
    revision: "1".to_string(),
    generated_at: now_rfc3339(),
    os: info.os.clone(),
    os_version: info.os_version.clone(),
    os_build: info.os_build.clone(),
    arch: info.arch.clone(),
    home_dir: info.home_dir.to_string_lossy().to_string(),
    devbox_root: info.devbox_root.to_string_lossy().to_string(),
    devbox_ready: info.devbox_ready,
    devbox_error: info.devbox_error.clone(),
    memory_total_bytes: info.memory_total_bytes,
    disk_total_bytes: info.disk_total_bytes,
    gpu_name: info.gpu_name.clone(),
    gpu_core_count: info.gpu_core_count,
  })
}
