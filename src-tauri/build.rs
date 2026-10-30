fn main() {
  let attributes =
    tauri_build::Attributes::new().app_manifest(tauri_build::AppManifest::new().commands(&[
      "get_environment",
      "get_settings",
      "update_settings",
      "test_proxy",
      "list_tools",
      "refresh_tools",
    ]));
  tauri_build::try_build(attributes).expect("failed to run tauri-build")
}
