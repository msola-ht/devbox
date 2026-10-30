use std::fs::{self, OpenOptions};
use std::io::{self, ErrorKind, Write};
use std::path::Path;

use serde::{Deserialize, Serialize};
use uuid::Uuid;

use super::{ManualProxySettings, ProxyProtocol, ProxySettings, Settings, SettingsError};
use crate::infrastructure::path::DevboxLayout;

const MIN_SCHEMA_VERSION: u32 = 1;
const SCHEMA_VERSION: u32 = 2;
const MAX_CONFIG_BYTES: u64 = 64 * 1024;

#[derive(Debug, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
struct StoredConfig {
  schema_version: u32,
  network: StoredNetwork,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct StoredVersion {
  schema_version: u32,
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
struct StoredNetwork {
  proxy: StoredProxySettings,
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(tag = "mode", rename_all = "camelCase")]
enum StoredProxySettings {
  System,
  Direct,
  Manual {
    protocol: StoredProxyProtocol,
    host: String,
    port: u16,
    #[serde(default)]
    bypass: Vec<String>,
  },
}

#[derive(Clone, Copy, Debug, Deserialize, Serialize)]
#[serde(rename_all = "lowercase")]
enum StoredProxyProtocol {
  Http,
  Socks5h,
}

pub fn load(layout: &DevboxLayout) -> Result<Option<Settings>, SettingsError> {
  let path = layout.config_file();
  match read_settings(&path) {
    Ok(Some(settings)) => Ok(Some(settings)),
    Ok(None) => read_backup(layout),
    Err(error) if matches!(error, SettingsError::UnsupportedVersion { .. }) => Err(error),
    Err(primary_error) => {
      log::warn!("主设置文件不可用，尝试读取备份: {primary_error}");
      match read_backup(layout) {
        Ok(Some(settings)) => Ok(Some(settings)),
        Ok(None) => Err(primary_error),
        Err(backup_error) => {
          log::error!("设置备份也无法恢复: {backup_error}");
          Err(primary_error)
        }
      }
    }
  }
}

pub fn save(layout: &DevboxLayout, settings: &Settings) -> Result<(), SettingsError> {
  let path = layout.config_file();
  let backup_path = layout.config_backup_file();
  let stored = StoredConfig::from(settings);
  let serialized =
    serde_json::to_vec_pretty(&stored).map_err(|error| SettingsError::WriteConfig {
      path: path.clone(),
      source: io::Error::new(ErrorKind::InvalidData, error),
    })?;

  match read_settings_bytes(&path) {
    Ok(Some((_, current))) => {
      write_atomic(&backup_path, &current).map_err(|source| SettingsError::WriteConfig {
        path: backup_path,
        source,
      })?;
    }
    Ok(None) => {}
    Err(error) => return Err(error),
  }

  write_atomic(&path, &serialized).map_err(|source| SettingsError::WriteConfig { path, source })
}

fn read_backup(layout: &DevboxLayout) -> Result<Option<Settings>, SettingsError> {
  let backup_path = layout.config_backup_file();
  let (settings, bytes) = match read_settings_bytes(&backup_path)? {
    Some(config) => config,
    None => return Ok(None),
  };

  let path = layout.config_file();
  log::warn!(
    "主设置文件不可用，正在从备份恢复: {}",
    backup_path.display()
  );
  write_atomic(&path, &bytes).map_err(|source| SettingsError::WriteConfig {
    path: path.clone(),
    source,
  })?;
  Ok(Some(settings))
}

fn read_settings(path: &Path) -> Result<Option<Settings>, SettingsError> {
  read_settings_bytes(path).map(|config| config.map(|(settings, _)| settings))
}

fn read_settings_bytes(path: &Path) -> Result<Option<(Settings, Vec<u8>)>, SettingsError> {
  let metadata = match fs::symlink_metadata(path) {
    Ok(metadata) => metadata,
    Err(error) if error.kind() == ErrorKind::NotFound => return Ok(None),
    Err(source) => {
      return Err(SettingsError::ReadConfig {
        path: path.to_path_buf(),
        source,
      });
    }
  };

  if metadata.file_type().is_symlink() || !metadata.is_file() {
    return Err(SettingsError::UnsafeConfigPath {
      path: path.to_path_buf(),
    });
  }
  if metadata.len() > MAX_CONFIG_BYTES {
    return Err(SettingsError::CorruptConfig {
      path: path.to_path_buf(),
    });
  }

  let bytes = fs::read(path).map_err(|source| SettingsError::ReadConfig {
    path: path.to_path_buf(),
    source,
  })?;
  if bytes.len() as u64 > MAX_CONFIG_BYTES {
    return Err(SettingsError::CorruptConfig {
      path: path.to_path_buf(),
    });
  }
  let settings = decode_settings(path, &bytes)?;
  Ok(Some((settings, bytes)))
}

fn decode_settings(path: &Path, bytes: &[u8]) -> Result<Settings, SettingsError> {
  let version: StoredVersion = serde_json::from_slice(bytes).map_err(|error| {
    log::warn!("设置文件解析失败 {}: {error}", path.display());
    SettingsError::CorruptConfig {
      path: path.to_path_buf(),
    }
  })?;

  if !(MIN_SCHEMA_VERSION..=SCHEMA_VERSION).contains(&version.schema_version) {
    return Err(SettingsError::UnsupportedVersion {
      version: version.schema_version,
    });
  }

  let stored: StoredConfig = serde_json::from_slice(bytes).map_err(|error| {
    log::warn!("设置文件格式解析失败 {}: {error}", path.display());
    SettingsError::CorruptConfig {
      path: path.to_path_buf(),
    }
  })?;

  let proxy = match stored.network.proxy {
    StoredProxySettings::System => ProxySettings::System,
    StoredProxySettings::Direct => ProxySettings::Direct,
    StoredProxySettings::Manual {
      protocol,
      host,
      port,
      bypass,
    } => ProxySettings::Manual(ManualProxySettings::try_new(
      protocol.into(),
      &host,
      port,
      bypass,
    )?),
  };

  Ok(Settings::new(proxy))
}

fn write_atomic(path: &Path, bytes: &[u8]) -> io::Result<()> {
  let parent = path
    .parent()
    .ok_or_else(|| io::Error::new(ErrorKind::InvalidInput, "设置文件缺少父目录"))?;
  let file_name = path
    .file_name()
    .and_then(|name| name.to_str())
    .unwrap_or("config.json");
  let temporary_path = parent.join(format!(".{file_name}.tmp-{}", Uuid::new_v4()));

  let result = (|| {
    let mut options = OpenOptions::new();
    options.create_new(true).write(true);
    #[cfg(unix)]
    {
      use std::os::unix::fs::OpenOptionsExt;
      options.mode(0o600);
    }
    let mut file = options.open(&temporary_path)?;
    file.write_all(bytes)?;
    restrict_file_permissions(&temporary_path)?;
    file.sync_all()?;
    fs::rename(&temporary_path, path)?;
    if let Err(error) = sync_parent_directory(parent) {
      log::warn!(
        "设置文件已替换，但父目录同步失败 {}: {error}",
        parent.display()
      );
    }
    Ok(())
  })();

  if result.is_err() {
    let _ = fs::remove_file(&temporary_path);
  }
  result
}

#[cfg(unix)]
fn sync_parent_directory(path: &Path) -> io::Result<()> {
  fs::File::open(path)?.sync_all()
}

#[cfg(not(unix))]
fn sync_parent_directory(_path: &Path) -> io::Result<()> {
  Ok(())
}

impl From<&Settings> for StoredConfig {
  fn from(settings: &Settings) -> Self {
    let proxy = match settings.proxy() {
      ProxySettings::System => StoredProxySettings::System,
      ProxySettings::Direct => StoredProxySettings::Direct,
      ProxySettings::Manual(proxy) => StoredProxySettings::Manual {
        protocol: proxy.protocol().into(),
        host: proxy.host().to_string(),
        port: proxy.port(),
        bypass: proxy.bypass().to_vec(),
      },
    };

    Self {
      schema_version: SCHEMA_VERSION,
      network: StoredNetwork { proxy },
    }
  }
}

impl From<StoredProxyProtocol> for ProxyProtocol {
  fn from(protocol: StoredProxyProtocol) -> Self {
    match protocol {
      StoredProxyProtocol::Http => Self::Http,
      StoredProxyProtocol::Socks5h => Self::Socks5h,
    }
  }
}

impl From<ProxyProtocol> for StoredProxyProtocol {
  fn from(protocol: ProxyProtocol) -> Self {
    match protocol {
      ProxyProtocol::Http => Self::Http,
      ProxyProtocol::Socks5h => Self::Socks5h,
    }
  }
}

#[cfg(unix)]
fn restrict_file_permissions(path: &Path) -> io::Result<()> {
  use std::os::unix::fs::PermissionsExt;

  let mut permissions = fs::metadata(path)?.permissions();
  permissions.set_mode(0o600);
  fs::set_permissions(path, permissions)
}

#[cfg(not(unix))]
fn restrict_file_permissions(_path: &Path) -> io::Result<()> {
  Ok(())
}
