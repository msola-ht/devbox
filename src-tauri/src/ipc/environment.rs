use serde::Serialize;
use tauri::State;

use super::error::IpcError;
use super::now_rfc3339;
use crate::modules::environment::{EnvironmentSnapshot, ShellEnvironmentStatus};
use crate::state::AppState;

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct EnvironmentSnapshotDto {
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
  default_shell: Option<String>,
  shell_environment_status: ShellEnvironmentStatusDto,
}

#[derive(Clone, Copy, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
enum ShellEnvironmentStatusDto {
  Ready,
  DefaultShellUnavailable,
  UnsupportedShell,
  LoginPathUnavailable,
}

#[tauri::command]
pub fn get_environment(state: State<'_, AppState>) -> Result<EnvironmentSnapshotDto, IpcError> {
  Ok(build_snapshot(&state, state.environment.snapshot()))
}

#[tauri::command]
pub async fn refresh_environment(
  state: State<'_, AppState>,
) -> Result<EnvironmentSnapshotDto, IpcError> {
  let snapshot = state.environment.refresh().await;
  Ok(build_snapshot(&state, snapshot))
}

fn build_snapshot(state: &AppState, snapshot: EnvironmentSnapshot) -> EnvironmentSnapshotDto {
  let info = snapshot.info;

  EnvironmentSnapshotDto {
    boot_id: state.boot_id.clone(),
    revision: snapshot.revision.to_string(),
    generated_at: now_rfc3339(),
    os: info.os,
    os_version: info.os_version,
    os_build: info.os_build,
    arch: info.arch,
    home_dir: info.home_dir.to_string_lossy().to_string(),
    devbox_root: info.devbox_root.to_string_lossy().to_string(),
    devbox_ready: info.devbox_ready,
    devbox_error: info.devbox_error,
    memory_total_bytes: info.memory_total_bytes,
    disk_total_bytes: info.disk_total_bytes,
    gpu_name: info.gpu_name,
    gpu_core_count: info.gpu_core_count,
    default_shell: info.default_shell,
    shell_environment_status: info.shell_environment_status.into(),
  }
}

impl From<ShellEnvironmentStatus> for ShellEnvironmentStatusDto {
  fn from(status: ShellEnvironmentStatus) -> Self {
    match status {
      ShellEnvironmentStatus::Ready => Self::Ready,
      ShellEnvironmentStatus::DefaultShellUnavailable => Self::DefaultShellUnavailable,
      ShellEnvironmentStatus::UnsupportedShell => Self::UnsupportedShell,
      ShellEnvironmentStatus::LoginPathUnavailable => Self::LoginPathUnavailable,
    }
  }
}
