mod infrastructure;
mod ipc;
mod modules;
mod state;

use infrastructure::path::{DevboxLayout, devbox_root, ensure_layout};
use modules::environment::{EnvironmentInfo, EnvironmentService, detect_system_profile};
use modules::settings::SettingsService;
use modules::toolchain::{
  BashPlugin, GitPlugin, NodePlugin, NpmPlugin, PluginRegistry, PnpmPlugin, PythonPlugin,
  ToolchainService, UvPlugin, ZshPlugin,
};
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
      let devbox_ready = devbox_error.is_none();

      if devbox_ready {
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
        "系统信息: version={:?} build={:?} memory_total={:?} disk_total={:?} gpu={:?} gpu_cores={:?} default_shell={:?} shell_environment_status={:?} login_shell_path_entries={}",
        profile.os_version,
        profile.os_build,
        profile.memory_total_bytes,
        profile.disk_total_bytes,
        profile.gpu_name,
        profile.gpu_core_count,
        profile.default_shell,
        profile.shell_environment_status,
        profile.login_shell_path.len()
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
        default_shell: profile.default_shell,
        shell_environment_status: profile.shell_environment_status,
        login_shell_path: profile.login_shell_path,
        home_dir: home_dir.clone(),
        devbox_root: root.clone(),
        devbox_ready,
        devbox_error,
      });
      let registry = PluginRegistry::new(vec![
        Box::new(UvPlugin),
        Box::new(PythonPlugin),
        Box::new(NodePlugin),
        Box::new(NpmPlugin),
        Box::new(PnpmPlugin),
        Box::new(GitPlugin),
        Box::new(BashPlugin),
        Box::new(ZshPlugin),
      ]);
      let layout = DevboxLayout::new(root);
      let settings = if !devbox_ready {
        SettingsService::unavailable("SETTINGS_UNAVAILABLE", "无法完成 DevBox 设置操作")
      } else {
        match SettingsService::new(layout.clone()) {
          Ok(settings) => settings,
          Err(error) => {
            log::error!("无法加载设置: {error}");
            SettingsService::unavailable(error.code(), error.user_message())
          }
        }
      };
      let toolchain = ToolchainService::new(registry, layout);
      app.manage(AppState::new(environment, settings, toolchain));

      Ok(())
    })
    .invoke_handler(tauri::generate_handler![
      ipc::environment::get_environment,
      ipc::environment::refresh_environment,
      ipc::settings::get_settings,
      ipc::settings::update_settings,
      ipc::settings::test_proxy,
      ipc::toolchain::list_tools,
      ipc::toolchain::refresh_tools,
    ])
    .run(tauri::generate_context!())
    .expect("error while building tauri application");
}
