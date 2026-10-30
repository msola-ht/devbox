use std::sync::mpsc::{self, Receiver, Sender};
use std::sync::{Arc, RwLock};
use std::thread;
use std::time::{Duration, Instant};

use super::{
  ProxySettings, Settings, SettingsError, SystemProxyEndpoint, SystemProxyInspection,
  SystemProxySource,
};
use crate::infrastructure::network::{
  NetworkClientError, NetworkRoute, ProxyEndpoint, SystemProxyEndpoint as DetectedProxyEndpoint,
  SystemProxyInspection as DetectedProxyInspection, SystemProxySource as DetectedProxySource,
  build_client, inspect_system_proxy as detect_system_proxy,
};
use crate::infrastructure::path::DevboxLayout;

const PROXY_TEST_URL: &str = "https://github.com/";
const PROXY_TEST_TIMEOUT: Duration = Duration::from_secs(12);

pub struct SettingsService {
  state: Arc<RwLock<SettingsState>>,
  update_sender: Option<Sender<UpdateCommand>>,
}

struct SettingsState {
  revision: u64,
  settings: Settings,
  load_error: Option<SettingsUnavailable>,
}

#[derive(Clone, Copy)]
struct SettingsUnavailable {
  code: &'static str,
  message: &'static str,
}

struct UpdateCommand {
  expected_revision: u64,
  proxy: ProxySettings,
  reply: tokio::sync::oneshot::Sender<Result<SettingsSnapshot, SettingsError>>,
}

#[derive(Clone, Debug)]
pub struct SettingsSnapshot {
  pub revision: u64,
  pub settings: Settings,
}

#[derive(Clone, Debug)]
pub struct ProxyTestResult {
  pub latency_ms: u64,
  pub http_status: u16,
}

impl SettingsService {
  pub fn new(layout: DevboxLayout) -> Result<Self, SettingsError> {
    let settings = super::repository::load(&layout)?.unwrap_or_default();
    let state = Arc::new(RwLock::new(SettingsState {
      revision: 1,
      settings,
      load_error: None,
    }));
    let writer_state = Arc::clone(&state);
    let (update_sender, update_receiver) = mpsc::channel();
    thread::Builder::new()
      .name("devbox-settings-writer".to_string())
      .spawn(move || run_writer(layout, writer_state, update_receiver))
      .map_err(|error| SettingsError::BlockingTaskFailed {
        message: error.to_string(),
      })?;

    Ok(Self {
      state,
      update_sender: Some(update_sender),
    })
  }

  pub fn unavailable(code: &'static str, message: &'static str) -> Self {
    Self {
      state: Arc::new(RwLock::new(SettingsState {
        revision: 1,
        settings: Settings::default(),
        load_error: Some(SettingsUnavailable { code, message }),
      })),
      update_sender: None,
    }
  }

  pub fn get(&self) -> Result<SettingsSnapshot, SettingsError> {
    let state = self
      .state
      .read()
      .map_err(|error| SettingsError::BlockingTaskFailed {
        message: error.to_string(),
      })?;
    if let Some(error) = state.load_error {
      return Err(SettingsError::Unavailable {
        code: error.code,
        message: error.message,
      });
    }
    Ok(SettingsSnapshot {
      revision: state.revision,
      settings: state.settings.clone(),
    })
  }

  pub async fn update(
    &self,
    expected_revision: u64,
    proxy: ProxySettings,
  ) -> Result<SettingsSnapshot, SettingsError> {
    let Some(sender) = &self.update_sender else {
      return Err(self.unavailable_error()?);
    };
    let (reply, response) = tokio::sync::oneshot::channel();
    sender
      .send(UpdateCommand {
        expected_revision,
        proxy,
        reply,
      })
      .map_err(|error| SettingsError::BlockingTaskFailed {
        message: error.to_string(),
      })?;

    response
      .await
      .map_err(|error| SettingsError::BlockingTaskFailed {
        message: error.to_string(),
      })?
  }

  pub async fn test_proxy(
    &self,
    proxy: &ProxySettings,
  ) -> Result<ProxyTestResult, NetworkClientError> {
    let route = route_for(proxy);
    let client = build_client(&route)?;
    let started_at = Instant::now();
    let response = tokio::time::timeout(PROXY_TEST_TIMEOUT, client.head(PROXY_TEST_URL).send())
      .await
      .map_err(|_| NetworkClientError::Timeout)?
      .map_err(classify_request_error)?;

    if response.status() == reqwest::StatusCode::PROXY_AUTHENTICATION_REQUIRED {
      return Err(NetworkClientError::ProxyAuthenticationRequired);
    }

    Ok(ProxyTestResult {
      latency_ms: started_at.elapsed().as_millis() as u64,
      http_status: response.status().as_u16(),
    })
  }

  pub async fn inspect_system_proxy(&self) -> Result<SystemProxyInspection, SettingsError> {
    let inspection = tokio::task::spawn_blocking(detect_system_proxy)
      .await
      .map_err(|error| SettingsError::BlockingTaskFailed {
        message: error.to_string(),
      })?;
    Ok(inspection.into())
  }

  fn unavailable_error(&self) -> Result<SettingsError, SettingsError> {
    let state = self
      .state
      .read()
      .map_err(|error| SettingsError::BlockingTaskFailed {
        message: error.to_string(),
      })?;
    let error = state.load_error.unwrap_or(SettingsUnavailable {
      code: "SETTINGS_UNAVAILABLE",
      message: "无法完成 DevBox 设置操作",
    });
    Ok(SettingsError::Unavailable {
      code: error.code,
      message: error.message,
    })
  }
}

impl From<DetectedProxyInspection> for SystemProxyInspection {
  fn from(inspection: DetectedProxyInspection) -> Self {
    Self::new(
      inspection.http.map(SystemProxyEndpoint::from),
      inspection.https.map(SystemProxyEndpoint::from),
    )
  }
}

impl From<DetectedProxyEndpoint> for SystemProxyEndpoint {
  fn from(endpoint: DetectedProxyEndpoint) -> Self {
    Self::new(
      endpoint.url,
      endpoint.source.into(),
      endpoint.requires_authentication,
    )
  }
}

impl From<DetectedProxySource> for SystemProxySource {
  fn from(source: DetectedProxySource) -> Self {
    match source {
      DetectedProxySource::Environment => Self::Environment,
      DetectedProxySource::OperatingSystem => Self::OperatingSystem,
    }
  }
}

fn classify_request_error(error: reqwest::Error) -> NetworkClientError {
  let message = error.to_string();
  if error.is_timeout() {
    NetworkClientError::Timeout
  } else if error.status() == Some(reqwest::StatusCode::PROXY_AUTHENTICATION_REQUIRED)
    || message
      .to_ascii_lowercase()
      .contains("proxy authorization required")
  {
    NetworkClientError::ProxyAuthenticationRequired
  } else {
    NetworkClientError::Request { message }
  }
}

fn run_writer(
  layout: DevboxLayout,
  state: Arc<RwLock<SettingsState>>,
  receiver: Receiver<UpdateCommand>,
) {
  while let Ok(command) = receiver.recv() {
    let result = apply_update(&layout, &state, command.expected_revision, command.proxy);
    let _ = command.reply.send(result);
  }
}

fn apply_update(
  layout: &DevboxLayout,
  state: &RwLock<SettingsState>,
  expected_revision: u64,
  proxy: ProxySettings,
) -> Result<SettingsSnapshot, SettingsError> {
  let next = Settings::new(proxy);
  let revision = {
    let state = state
      .read()
      .map_err(|error| SettingsError::BlockingTaskFailed {
        message: error.to_string(),
      })?;
    if let Some(error) = state.load_error {
      return Err(SettingsError::Unavailable {
        code: error.code,
        message: error.message,
      });
    }
    if state.revision != expected_revision {
      return Err(SettingsError::Conflict {
        expected: expected_revision,
        actual: state.revision,
      });
    }
    if next == state.settings {
      return Ok(SettingsSnapshot {
        revision: state.revision,
        settings: state.settings.clone(),
      });
    }
    state.revision
  };

  super::repository::save(layout, &next)?;

  let mut state = state
    .write()
    .map_err(|error| SettingsError::BlockingTaskFailed {
      message: error.to_string(),
    })?;
  state.settings = next;
  state.revision = revision + 1;
  Ok(SettingsSnapshot {
    revision: state.revision,
    settings: state.settings.clone(),
  })
}

fn route_for(proxy: &ProxySettings) -> NetworkRoute {
  match proxy {
    ProxySettings::System => NetworkRoute::system(),
    ProxySettings::Direct => NetworkRoute::direct(),
    ProxySettings::Manual(proxy) => NetworkRoute::manual(ProxyEndpoint {
      url: proxy.endpoint_url(),
      bypass: proxy.bypass().to_vec(),
    }),
  }
}
