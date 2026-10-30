use std::str::FromStr;

use serde::{Deserialize, Serialize};
use tauri::State;

use super::error::IpcError;
use super::now_rfc3339;
use crate::infrastructure::network::NetworkClientError;
use crate::modules::settings::{
  ManualProxySettings, ProxyProtocol, ProxySettings, SettingsError, SettingsSnapshot,
  SystemProxyEndpoint, SystemProxyInspection, SystemProxySource,
};
use crate::state::AppState;

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(tag = "mode", rename_all = "camelCase")]
pub enum ProxySettingsDto {
  System,
  Direct,
  Manual {
    protocol: ProxyProtocolDto,
    host: String,
    port: u16,
    #[serde(default)]
    bypass: Vec<String>,
  },
}

#[derive(Clone, Copy, Debug, Deserialize, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum ProxyProtocolDto {
  Http,
  Socks5h,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UpdateSettingsRequest {
  expected_revision: String,
  proxy: ProxySettingsDto,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProxyTestRequest {
  proxy: ProxySettingsDto,
}

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SettingsSnapshotDto {
  boot_id: String,
  revision: String,
  generated_at: String,
  proxy: ProxySettingsDto,
  system_proxy: SystemProxyDto,
}

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ProxyTestResultDto {
  checked_at: String,
  latency_ms: u64,
  http_status: u16,
}

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
struct SystemProxyDto {
  checked_at: String,
  http: Option<SystemProxyEndpointDto>,
  https: Option<SystemProxyEndpointDto>,
}

#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
struct SystemProxyEndpointDto {
  url: String,
  source: SystemProxySourceDto,
  requires_authentication: bool,
}

#[derive(Clone, Copy, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
enum SystemProxySourceDto {
  Environment,
  OperatingSystem,
}

#[tauri::command]
pub async fn get_settings(state: State<'_, AppState>) -> Result<SettingsSnapshotDto, IpcError> {
  let snapshot = state.settings.get().map_err(map_settings_error)?;
  let system_proxy = state
    .settings
    .inspect_system_proxy()
    .await
    .map_err(map_settings_error)?;
  Ok(build_snapshot(&state, &snapshot, system_proxy))
}

#[tauri::command]
pub async fn update_settings(
  state: State<'_, AppState>,
  request: UpdateSettingsRequest,
) -> Result<SettingsSnapshotDto, IpcError> {
  let expected_revision = u64::from_str(&request.expected_revision)
    .map_err(|_| IpcError::new("INVALID_INPUT", "设置版本号格式无效"))?;
  let proxy = request.proxy.into_domain()?;
  let system_proxy = state
    .settings
    .inspect_system_proxy()
    .await
    .map_err(map_settings_error)?;
  let snapshot = state
    .settings
    .update(expected_revision, proxy)
    .await
    .map_err(map_settings_error)?;
  Ok(build_snapshot(&state, &snapshot, system_proxy))
}

#[tauri::command]
pub async fn test_proxy(
  state: State<'_, AppState>,
  request: ProxyTestRequest,
) -> Result<ProxyTestResultDto, IpcError> {
  let proxy = request.proxy.into_domain()?;
  let result = state
    .settings
    .test_proxy(&proxy)
    .await
    .map_err(map_network_error)?;
  Ok(ProxyTestResultDto {
    checked_at: now_rfc3339(),
    latency_ms: result.latency_ms,
    http_status: result.http_status,
  })
}

fn build_snapshot(
  state: &AppState,
  snapshot: &SettingsSnapshot,
  system_proxy: SystemProxyInspection,
) -> SettingsSnapshotDto {
  SettingsSnapshotDto {
    boot_id: state.boot_id.clone(),
    revision: snapshot.revision.to_string(),
    generated_at: now_rfc3339(),
    proxy: ProxySettingsDto::from(snapshot.settings.proxy()),
    system_proxy: SystemProxyDto {
      checked_at: now_rfc3339(),
      http: system_proxy
        .http()
        .cloned()
        .map(SystemProxyEndpointDto::from),
      https: system_proxy
        .https()
        .cloned()
        .map(SystemProxyEndpointDto::from),
    },
  }
}

impl From<SystemProxyEndpoint> for SystemProxyEndpointDto {
  fn from(endpoint: SystemProxyEndpoint) -> Self {
    Self {
      url: endpoint.url().to_string(),
      source: endpoint.source().into(),
      requires_authentication: endpoint.requires_authentication(),
    }
  }
}

impl From<SystemProxySource> for SystemProxySourceDto {
  fn from(source: SystemProxySource) -> Self {
    match source {
      SystemProxySource::Environment => Self::Environment,
      SystemProxySource::OperatingSystem => Self::OperatingSystem,
    }
  }
}

impl ProxySettingsDto {
  fn into_domain(self) -> Result<ProxySettings, IpcError> {
    match self {
      Self::System => Ok(ProxySettings::System),
      Self::Direct => Ok(ProxySettings::Direct),
      Self::Manual {
        protocol,
        host,
        port,
        bypass,
      } => ManualProxySettings::try_new(protocol.into(), &host, port, bypass)
        .map(ProxySettings::Manual)
        .map_err(map_settings_error),
    }
  }
}

impl From<&ProxySettings> for ProxySettingsDto {
  fn from(proxy: &ProxySettings) -> Self {
    match proxy {
      ProxySettings::System => Self::System,
      ProxySettings::Direct => Self::Direct,
      ProxySettings::Manual(proxy) => Self::Manual {
        protocol: proxy.protocol().into(),
        host: proxy.host().to_string(),
        port: proxy.port(),
        bypass: proxy.bypass().to_vec(),
      },
    }
  }
}

impl From<ProxyProtocolDto> for ProxyProtocol {
  fn from(protocol: ProxyProtocolDto) -> Self {
    match protocol {
      ProxyProtocolDto::Http => Self::Http,
      ProxyProtocolDto::Socks5h => Self::Socks5h,
    }
  }
}

impl From<ProxyProtocol> for ProxyProtocolDto {
  fn from(protocol: ProxyProtocol) -> Self {
    match protocol {
      ProxyProtocol::Http => Self::Http,
      ProxyProtocol::Socks5h => Self::Socks5h,
    }
  }
}

fn map_settings_error(error: SettingsError) -> IpcError {
  log::error!("设置操作失败: {error}");
  IpcError::new(error.code(), error.user_message())
}

fn map_network_error(error: NetworkClientError) -> IpcError {
  log::error!("代理连接测试失败: {error}");
  match error {
    NetworkClientError::InvalidProxy { .. } => {
      IpcError::new("PROXY_CONFIGURATION_INVALID", "代理配置无效")
    }
    NetworkClientError::Timeout => IpcError::new("PROXY_CONNECTION_FAILED", "代理连接超时"),
    NetworkClientError::ProxyAuthenticationRequired => IpcError::new(
      "PROXY_AUTH_UNSUPPORTED",
      "代理要求身份验证，当前版本不支持代理账号密码",
    ),
    NetworkClientError::Build { .. } | NetworkClientError::Request { .. } => {
      IpcError::new("PROXY_CONNECTION_FAILED", "无法通过当前代理访问网络")
    }
  }
}
