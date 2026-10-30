pub mod error;
mod model;
mod repository;
mod service;

pub use error::SettingsError;
pub use model::{
  ManualProxySettings, ProxyProtocol, ProxySettings, Settings, SystemProxyEndpoint,
  SystemProxyInspection, SystemProxySource,
};
pub use service::{SettingsService, SettingsSnapshot};
