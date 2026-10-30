mod infrastructure;
mod ipc;
mod modules;
mod state;

use infrastructure::path::{DevboxLayout, devbox_root, ensure_layout};
use modules::environment::{
  EnvironmentInfo, EnvironmentService, detect_system_profile,
};
use modules::toolchain::{PluginRegistry, PythonPlugin, ToolchainService, UvPlugin};
use state::AppState;
use tauri::Manager;
use tauri_plugin_log::{Target, TargetKind};

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
  tauri::Builder::default()
    .setup(|app| {
      let home_dir = app.path().home_dir()?;
      let root = devbox_root(&home_dir);
      let devbox_error = ensure_layout(&root)
        .err()
        .map(|error| format!("无法初始化 .devbox 目录: {error}"));

      if devbox_error.is_none() {
        let mut targets = vec![Target::new(TargetKind::Folder {
          path: root.join("logs"),
          file_name: Some("devbox".to_string()),
        })];
        #[cfg(debug_assertions)]
        targets.push(Target::new(TargetKind::Stdout));

        app.handle().plugin(
          tauri_plugin_log::Builder::default()
            .level(log::LevelFilter::Info)
            .max_file_size(10 * 1024 * 1024)
            .targets(targets)
            .build(),
        )?;
      }

      let profile = tauri::async_runtime::block_on(detect_system_profile(&home_dir));
      log::info!(
        "系统信息: version={:?} build={:?} memory_total={:?} disk_total={:?} gpu={:?} gpu_cores={:?}",
        profile.os_version,
        profile.os_build,
        profile.memory_total_bytes,
        profile.disk_total_bytes,
        profile.gpu_name,
        profile.gpu_core_count
      );
      let environment = EnvironmentService::new(EnvironmentInfo {
        os: std::env::consts::OS.to_string(),
        os_version: profile.os_version,
        os_build: profile.os_build,
        arch: std::env::consts::ARCH.to_string(),
        memory_total_bytes: profile.memory_total_bytes,
        disk_total_bytes: profile.disk_total_bytes,
        gpu_name: profile.gpu_name,
        gpu_core_count: profile.gpu_core_count,
        home_dir: home_dir.clone(),
        devbox_root: root.clone(),
        devbox_ready: devbox_error.is_none(),
        devbox_error,
      });
      let registry = PluginRegistry::new(vec![
        Box::new(UvPlugin),
        Box::new(PythonPlugin),
      ]);
      let toolchain = ToolchainService::new(registry, DevboxLayout::new(root));
      app.manage(AppState::new(environment, toolchain));

      Ok(())
    })
    .invoke_handler(tauri::generate_handler![
      ipc::environment::get_environment,
      ipc::toolchain::list_tools,
      ipc::toolchain::refresh_tools,
    ])
    .run(tauri::generate_context!())
    .expect("error while building tauri application");
}
