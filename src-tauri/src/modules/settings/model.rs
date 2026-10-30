use std::net::{IpAddr, Ipv4Addr, Ipv6Addr};

use super::SettingsError;

const MAX_HOST_LENGTH: usize = 253;
const MAX_BYPASS_ENTRIES: usize = 64;
const MAX_BYPASS_ENTRY_LENGTH: usize = 255;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ProxyProtocol {
  Http,
  Socks5h,
}

impl ProxyProtocol {
  pub fn as_str(self) -> &'static str {
    match self {
      Self::Http => "http",
      Self::Socks5h => "socks5h",
    }
  }
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ManualProxySettings {
  protocol: ProxyProtocol,
  host: String,
  port: u16,
  bypass: Vec<String>,
}

impl ManualProxySettings {
  pub fn try_new(
    protocol: ProxyProtocol,
    host: &str,
    port: u16,
    bypass: Vec<String>,
  ) -> Result<Self, SettingsError> {
    if port == 0 {
      return Err(SettingsError::InvalidPort);
    }

    Ok(Self {
      protocol,
      host: normalize_host(host)?,
      port,
      bypass: normalize_bypass(bypass)?,
    })
  }

  pub fn protocol(&self) -> ProxyProtocol {
    self.protocol
  }

  pub fn host(&self) -> &str {
    &self.host
  }

  pub fn port(&self) -> u16 {
    self.port
  }

  pub fn bypass(&self) -> &[String] {
    &self.bypass
  }

  pub fn endpoint_url(&self) -> String {
    format!("{}://{}:{}", self.protocol.as_str(), self.host, self.port)
  }
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum ProxySettings {
  System,
  Direct,
  Manual(ManualProxySettings),
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Settings {
  proxy: ProxySettings,
}

impl Settings {
  pub fn new(proxy: ProxySettings) -> Self {
    Self { proxy }
  }

  pub fn proxy(&self) -> &ProxySettings {
    &self.proxy
  }
}

impl Default for Settings {
  fn default() -> Self {
    Self {
      proxy: ProxySettings::System,
    }
  }
}

fn normalize_host(input: &str) -> Result<String, SettingsError> {
  let host = input.trim().trim_end_matches('.');
  if host.is_empty()
    || host.len() > MAX_HOST_LENGTH
    || host
      .chars()
      .any(|character| character.is_control() || character.is_whitespace())
  {
    return Err(SettingsError::InvalidHost);
  }

  if let Ok(address) = host.parse::<Ipv6Addr>() {
    return Ok(format!("[{address}]"));
  }

  if host.starts_with('[') && host.ends_with(']') {
    let inner = &host[1..host.len() - 1];
    return inner
      .parse::<Ipv6Addr>()
      .map(|address| format!("[{address}]"))
      .map_err(|_| SettingsError::InvalidHost);
  }

  if let Ok(address) = host.parse::<Ipv4Addr>() {
    return Ok(address.to_string());
  }

  if host.split('.').any(|label| {
    label.is_empty()
      || label.len() > 63
      || label.starts_with('-')
      || label.ends_with('-')
      || !label
        .chars()
        .all(|character| character.is_ascii_alphanumeric() || character == '-')
  }) {
    return Err(SettingsError::InvalidHost);
  }

  Ok(host.to_ascii_lowercase())
}

fn normalize_bypass(entries: Vec<String>) -> Result<Vec<String>, SettingsError> {
  if entries.len() > MAX_BYPASS_ENTRIES {
    return Err(SettingsError::InvalidBypass);
  }

  let mut normalized = Vec::new();
  for entry in entries {
    let entry = entry.trim();
    if entry.is_empty() {
      continue;
    }
    let entry = normalize_bypass_entry(entry)?;
    if !normalized.iter().any(|existing| existing == &entry) {
      normalized.push(entry);
    }
  }

  Ok(normalized)
}

fn normalize_bypass_entry(entry: &str) -> Result<String, SettingsError> {
  if entry.len() > MAX_BYPASS_ENTRY_LENGTH
    || entry.contains("://")
    || entry
      .chars()
      .any(|character| character.is_control() || character.is_whitespace())
  {
    return Err(SettingsError::InvalidBypass);
  }

  if entry == "*" {
    return Ok(entry.to_string());
  }

  if let Some((address, prefix)) = entry.split_once('/') {
    if prefix.contains('/') {
      return Err(SettingsError::InvalidBypass);
    }
    let address = parse_ip_address(address)?;
    let prefix = prefix
      .parse::<u8>()
      .map_err(|_| SettingsError::InvalidBypass)?;
    let max_prefix = match address {
      IpAddr::V4(_) => 32,
      IpAddr::V6(_) => 128,
    };
    if prefix > max_prefix {
      return Err(SettingsError::InvalidBypass);
    }
    return Ok(format!("{address}/{prefix}"));
  }

  if let Ok(address) = parse_ip_address(entry) {
    return Ok(address.to_string());
  }

  normalize_bypass_domain(entry)
}

fn parse_ip_address(input: &str) -> Result<IpAddr, SettingsError> {
  let input = input
    .strip_prefix('[')
    .and_then(|address| address.strip_suffix(']'))
    .unwrap_or(input);
  input
    .parse::<IpAddr>()
    .map_err(|_| SettingsError::InvalidBypass)
}

fn normalize_bypass_domain(input: &str) -> Result<String, SettingsError> {
  let domain = input
    .strip_prefix('.')
    .unwrap_or(input)
    .trim_end_matches('.');
  if domain.is_empty()
    || domain.len() > MAX_HOST_LENGTH
    || domain.split('.').any(|label| {
      label.is_empty()
        || label.len() > 63
        || label.starts_with('-')
        || label.ends_with('-')
        || !label
          .chars()
          .all(|character| character.is_ascii_alphanumeric() || character == '-')
    })
  {
    return Err(SettingsError::InvalidBypass);
  }

  Ok(domain.to_ascii_lowercase())
}
