use std::collections::BTreeMap;
use std::fs;
use std::io::ErrorKind;
use std::path::{Component, Path, PathBuf};
use std::sync::RwLock;
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::Duration;

use serde::Deserialize;

use super::error::ToolchainError;
use super::model::{
  InstallationSource, InstallationSummary, ToolDescriptor, ToolSummary, ToolchainSnapshot,
};
use super::plugin::{ToolPlugin, ToolPluginContext};
use super::plugins::PluginRegistry;
use crate::infrastructure::path::DevboxLayout;
use crate::infrastructure::process::probe_version;
use crate::modules::environment::EnvironmentInfo;

const VERSION_PROBE_TIMEOUT: Duration = Duration::from_secs(3);

pub struct ToolchainService {
  registry: PluginRegistry,
  layout: DevboxLayout,
  revision: AtomicU64,
  cache: RwLock<Option<CachedTools>>,
}

#[derive(Clone)]
struct CachedTools {
  revision: u64,
  tools: Vec<ToolSummary>,
}

impl ToolchainService {
  pub fn new(registry: PluginRegistry, layout: DevboxLayout) -> Self {
    Self {
      registry,
      layout,
      revision: AtomicU64::new(0),
      cache: RwLock::new(None),
    }
  }

  pub async fn list_tools(
    &self,
    environment: &EnvironmentInfo,
  ) -> Result<ToolchainSnapshot, ToolchainError> {
    if let Some(cached) = self.cached_tools() {
      return Ok(ToolchainSnapshot {
        revision: cached.revision,
        tools: cached.tools,
      });
    }

    self.refresh_tools(environment).await
  }

  pub async fn refresh_tools(
    &self,
    environment: &EnvironmentInfo,
  ) -> Result<ToolchainSnapshot, ToolchainError> {
    let revision = self.next_revision();
    let tools = self.detect_tools(environment).await?;
    self.cache_tools(revision, tools.clone());

    let summary = tools
      .iter()
      .map(|tool| {
        format!(
          "{} ({} system, {} managed)",
          tool.tool_id,
          tool.system_installations.len(),
          tool.managed_installations.len()
        )
      })
      .collect::<Vec<_>>()
      .join(", ");
    log::info!("工具检测完成 revision={revision}: {summary}");

    Ok(ToolchainSnapshot { revision, tools })
  }

  fn next_revision(&self) -> u64 {
    self.revision.fetch_add(1, Ordering::SeqCst) + 1
  }

  fn cache_tools(&self, revision: u64, tools: Vec<ToolSummary>) {
    if let Ok(mut cache) = self.cache.write() {
      let is_newer = cache
        .as_ref()
        .is_none_or(|cached| revision > cached.revision);
      if is_newer {
        *cache = Some(CachedTools { revision, tools });
      }
    }
  }

  fn cached_tools(&self) -> Option<CachedTools> {
    self.cache.read().ok().and_then(|cache| cache.clone())
  }

  async fn detect_tools(
    &self,
    environment: &EnvironmentInfo,
  ) -> Result<Vec<ToolSummary>, ToolchainError> {
    let mut tools = Vec::new();
    let context = ToolPluginContext {
      os: &environment.os,
      home_dir: &environment.home_dir,
    };

    for plugin in self.registry.plugins() {
      let mut summary = ToolSummary::empty(plugin.descriptor());
      summary.system_installations = self.detect_system_installations(plugin, &context).await?;
      let (managed_installations, active_installation_id) =
        self.scan_managed_state(plugin).await?;
      summary.managed_installations = managed_installations;
      summary.active_installation_id = active_installation_id.filter(|active| {
        summary
          .managed_installations
          .iter()
          .any(|installation| installation.installation_id == *active && installation.healthy)
      });
      tools.push(summary);
    }

    Ok(tools)
  }

  async fn detect_system_installations(
    &self,
    plugin: &dyn ToolPlugin,
    environment: &ToolPluginContext<'_>,
  ) -> Result<Vec<InstallationSummary>, ToolchainError> {
    let mut installations = Vec::new();

    for (canonical, display_path) in self.system_candidates(plugin, environment).await? {
      let source_hint = plugin.source_hint(Path::new(&canonical));
      let probe = probe_version(&display_path, plugin.version_args(), VERSION_PROBE_TIMEOUT).await;

      match probe {
        Ok(output) => {
          let version = plugin.parse_version_output(&output);
          let healthy = version.is_some();
          installations.push(InstallationSummary {
            installation_id: format!("system:{canonical}"),
            version,
            path: display_path.to_string_lossy().to_string(),
            source: InstallationSource::System,
            source_hint,
            healthy,
            error_code: (!healthy).then(|| "VERSION_OUTPUT_UNRECOGNIZED".to_string()),
          });
        }
        Err(error) => {
          log::warn!(
            "{} system probe failed for {}: {error}",
            plugin.descriptor().tool_id,
            display_path.display()
          );
          installations.push(InstallationSummary {
            installation_id: format!("system:{canonical}"),
            version: None,
            path: display_path.to_string_lossy().to_string(),
            source: InstallationSource::System,
            source_hint,
            healthy: false,
            error_code: Some("SYSTEM_PROBE_FAILED".to_string()),
          });
        }
      }
    }

    Ok(installations)
  }

  async fn system_candidates(
    &self,
    plugin: &dyn ToolPlugin,
    environment: &ToolPluginContext<'_>,
  ) -> Result<Vec<(String, PathBuf)>, ToolchainError> {
    let mut candidates = BTreeMap::new();
    let mut directories = Vec::new();

    if let Some(path) = std::env::var_os("PATH") {
      directories.extend(std::env::split_paths(&path));
    }
    directories.extend(plugin.known_directories(environment));
    let executables = plugin.executable_names().to_vec();

    tokio::task::spawn_blocking(move || {
      for directory in directories {
        if !directory.is_absolute() {
          continue;
        }

        for executable in &executables {
          let candidate = directory.join(executable);
          if !candidate.is_file() {
            continue;
          }

          let canonical = fs::canonicalize(&candidate)
            .unwrap_or_else(|_| candidate.clone())
            .to_string_lossy()
            .to_string();
          candidates.entry(canonical).or_insert(candidate);
        }
      }

      candidates.into_iter().collect()
    })
    .await
    .map_err(|error| ToolchainError::BlockingTaskFailed {
      operation: "系统安装候选发现",
      message: error.to_string(),
    })
  }

  async fn scan_managed_state(
    &self,
    plugin: &dyn ToolPlugin,
  ) -> Result<(Vec<InstallationSummary>, Option<String>), ToolchainError> {
    let layout = self.layout.clone();
    let descriptor = plugin.descriptor().clone();

    tokio::task::spawn_blocking(move || {
      let installations = scan_managed_installations(&layout, &descriptor)?;
      let active = read_active_installation(&layout, descriptor.tool_id);
      Ok((installations, active))
    })
    .await
    .map_err(|error| ToolchainError::BlockingTaskFailed {
      operation: "托管安装扫描",
      message: error.to_string(),
    })?
  }
}

fn scan_managed_installations(
  layout: &DevboxLayout,
  descriptor: &ToolDescriptor,
) -> Result<Vec<InstallationSummary>, ToolchainError> {
  let tool_root = layout.managed_tool_root(descriptor.tool_id);

  let tool_root_metadata = match fs::symlink_metadata(&tool_root) {
    Ok(metadata) => metadata,
    Err(error) if error.kind() == ErrorKind::NotFound => return Ok(Vec::new()),
    Err(source) => {
      return Err(ToolchainError::ReadManagedToolRoot {
        path: tool_root,
        source,
      });
    }
  };
  if tool_root_metadata.file_type().is_symlink() || !tool_root_metadata.is_dir() {
    return Err(ToolchainError::UnsafeManagedToolPath { path: tool_root });
  }

  let entries = fs::read_dir(&tool_root).map_err(|source| {
    ToolchainError::ReadManagedToolEntries {
      path: tool_root.clone(),
      source,
    }
  })?;
  let canonical_tool_root = fs::canonicalize(&tool_root).map_err(|source| {
    ToolchainError::CanonicalizeManagedToolRoot {
      path: tool_root.clone(),
      source,
    }
  })?;

  let mut installations = Vec::new();
  for entry in entries {
    let entry = entry.map_err(|source| ToolchainError::ReadManagedToolEntry { source })?;
    let entry_type = entry
      .file_type()
      .map_err(|source| ToolchainError::ReadManagedToolEntryType { source })?;
    if !entry_type.is_dir() {
      continue;
    }
    let install_path = entry.path();

    installations.push(read_managed_installation(
      descriptor,
      &install_path,
      &canonical_tool_root,
    ));
  }

  Ok(installations)
}

fn read_active_installation(layout: &DevboxLayout, tool_id: &str) -> Option<String> {
  let active_path = layout.active_installation_pointer(tool_id);
  let contents = match fs::read_to_string(&active_path) {
    Ok(contents) => contents,
    Err(error) if error.kind() == ErrorKind::NotFound => return None,
    Err(error) => {
      log::warn!("无法读取 active 指针 {}: {error}", active_path.display());
      return None;
    }
  };

  match serde_json::from_str::<ActiveInstallationPointer>(&contents) {
    Ok(pointer)
      if pointer.schema_version == 1
        && pointer.tool_id == tool_id
        && !pointer.installation_id.trim().is_empty() =>
    {
      Some(pointer.installation_id)
    }
    Err(error) => {
      log::warn!("active 指针格式无效 {}: {error}", active_path.display());
      None
    }
    Ok(_) => {
      log::warn!("active 指针字段不匹配 {}", active_path.display());
      None
    }
  }
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct ManagedInstallationManifest {
  schema_version: u32,
  tool_id: String,
  plugin_contract_version: u32,
  installation_id: String,
  version_id: String,
  executable_relative_path: String,
  healthy: bool,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct ActiveInstallationPointer {
  schema_version: u32,
  tool_id: String,
  installation_id: String,
}

fn read_managed_installation(
  descriptor: &ToolDescriptor,
  install_path: &Path,
  canonical_tool_root: &Path,
) -> InstallationSummary {
  let install_id = format!(
    "managed:{}",
    install_path
      .file_name()
      .map(|name| name.to_string_lossy().to_string())
      .unwrap_or_else(|| "unknown".to_string())
  );
  let manifest_path = install_path.join("install.json");
  let canonical_install_path = match fs::canonicalize(install_path) {
    Ok(path) if path.starts_with(canonical_tool_root) => path,
    _ => return damaged_managed_installation(install_path, install_id),
  };

  let manifest = match fs::read_to_string(&manifest_path)
    .map_err(|error| error.to_string())
    .and_then(|contents| {
      serde_json::from_str::<ManagedInstallationManifest>(&contents)
        .map_err(|error| error.to_string())
    }) {
    Ok(manifest) => manifest,
    Err(error) => {
      log::warn!("托管安装清单无效 {}: {error}", manifest_path.display());
      return damaged_managed_installation(install_path, install_id);
    }
  };

  if manifest.schema_version != 1
    || manifest.tool_id != descriptor.tool_id
    || manifest.plugin_contract_version != descriptor.contract_version
    || manifest.installation_id.trim().is_empty()
    || manifest.version_id.trim().is_empty()
    || !manifest.healthy
  {
    return damaged_managed_installation(install_path, install_id);
  }

  let executable_relative_path = Path::new(&manifest.executable_relative_path);
  if !is_safe_relative_path(executable_relative_path) {
    return damaged_managed_installation(install_path, install_id);
  }

  let executable = canonical_install_path.join(executable_relative_path);
  let executable_inside_installation = fs::canonicalize(&executable)
    .is_ok_and(|path| path.starts_with(&canonical_install_path) && path.is_file());
  if !executable_inside_installation {
    return damaged_managed_installation(install_path, install_id);
  }

  InstallationSummary {
    installation_id: manifest.installation_id,
    version: Some(manifest.version_id),
    path: install_path.to_string_lossy().to_string(),
    source: InstallationSource::Managed,
    source_hint: Some("DevBox".to_string()),
    healthy: true,
    error_code: None,
  }
}

fn is_safe_relative_path(path: &Path) -> bool {
  !path.as_os_str().is_empty()
    && path
      .components()
      .all(|component| matches!(component, Component::Normal(_)))
}

fn damaged_managed_installation(install_path: &Path, install_id: String) -> InstallationSummary {
  InstallationSummary {
    installation_id: install_id,
    version: None,
    path: install_path.to_string_lossy().to_string(),
    source: InstallationSource::Managed,
    source_hint: Some("DevBox".to_string()),
    healthy: false,
    error_code: Some("INSTALLATION_DAMAGED".to_string()),
  }
}
