use serde::Serialize;
use tauri::State;

use super::error::IpcError;
use super::now_rfc3339;
use crate::modules::toolchain::{
  InstallationSource, InstallationSummary, ToolSummary, ToolchainError, ToolchainSnapshot,
};
use crate::state::AppState;

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ToolListSnapshot {
  boot_id: String,
  revision: String,
  generated_at: String,
  tools: Vec<ToolSummaryDto>,
}

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
struct ToolSummaryDto {
  tool_id: String,
  display_name: String,
  description: String,
  plugin_contract_version: String,
  system_installations: Vec<InstallationSummaryDto>,
  managed_installations: Vec<InstallationSummaryDto>,
  active_installation_id: Option<String>,
  current_installation_id: Option<String>,
}

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
struct InstallationSummaryDto {
  installation_id: String,
  version: Option<String>,
  path: String,
  source: InstallationSourceDto,
  source_hint: Option<String>,
  healthy: bool,
  error_code: Option<String>,
}

#[derive(Clone, Copy, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
enum InstallationSourceDto {
  System,
  Managed,
}

#[tauri::command]
pub async fn list_tools(state: State<'_, AppState>) -> Result<ToolListSnapshot, IpcError> {
  let environment = state.environment.info().clone();
  let snapshot = state
    .toolchain
    .list_tools(&environment)
    .await
    .map_err(map_toolchain_error)?;
  Ok(build_snapshot(&state, &snapshot))
}

#[tauri::command]
pub async fn refresh_tools(state: State<'_, AppState>) -> Result<ToolListSnapshot, IpcError> {
  let environment = state.environment.info().clone();
  let snapshot = state
    .toolchain
    .refresh_tools(&environment)
    .await
    .map_err(map_toolchain_error)?;
  Ok(build_snapshot(&state, &snapshot))
}

fn build_snapshot(state: &AppState, snapshot: &ToolchainSnapshot) -> ToolListSnapshot {
  ToolListSnapshot {
    boot_id: state.boot_id.clone(),
    revision: snapshot.revision.to_string(),
    generated_at: now_rfc3339(),
    tools: snapshot.tools.iter().map(ToolSummaryDto::from).collect(),
  }
}

impl From<&ToolSummary> for ToolSummaryDto {
  fn from(tool: &ToolSummary) -> Self {
    Self {
      tool_id: tool.tool_id.clone(),
      display_name: tool.display_name.clone(),
      description: tool.description.clone(),
      plugin_contract_version: tool.plugin_contract_version.clone(),
      system_installations: tool
        .system_installations
        .iter()
        .map(InstallationSummaryDto::from)
        .collect(),
      managed_installations: tool
        .managed_installations
        .iter()
        .map(InstallationSummaryDto::from)
        .collect(),
      active_installation_id: tool.active_installation_id.clone(),
      current_installation_id: tool.current_installation_id.clone(),
    }
  }
}

impl From<&InstallationSummary> for InstallationSummaryDto {
  fn from(installation: &InstallationSummary) -> Self {
    Self {
      installation_id: installation.installation_id.clone(),
      version: installation.version.clone(),
      path: installation.path.clone(),
      source: match installation.source {
        InstallationSource::System => InstallationSourceDto::System,
        InstallationSource::Managed => InstallationSourceDto::Managed,
      },
      source_hint: installation.source_hint.clone(),
      healthy: installation.healthy,
      error_code: installation.error_code.clone(),
    }
  }
}

fn map_toolchain_error(error: ToolchainError) -> IpcError {
  log::error!("工具检测失败: {error}");
  let (code, message) = match error {
    ToolchainError::UnsafeManagedToolPath { .. } => (
      "TOOLS_DIRECTORY_UNSAFE",
      "DevBox 托管工具目录不安全，已停止检测",
    ),
    ToolchainError::ReadManagedToolRoot { .. }
    | ToolchainError::ReadManagedToolEntries { .. }
    | ToolchainError::ReadManagedToolEntry { .. }
    | ToolchainError::ReadManagedToolEntryType { .. }
    | ToolchainError::CanonicalizeManagedToolRoot { .. } => (
      "TOOLS_DIRECTORY_UNAVAILABLE",
      "无法读取 DevBox 托管工具目录",
    ),
    ToolchainError::BlockingTaskFailed { .. } => {
      ("TOOLS_SCAN_FAILED", "无法完成本机工具检测")
    }
  };
  IpcError::new(code, message)
}
