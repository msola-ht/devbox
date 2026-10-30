fn main() {
  let attributes =
    tauri_build::Attributes::new().app_manifest(tauri_build::AppManifest::new().commands(&[
      "get_environment",
      "list_tools",
      "refresh_tools",
    ]));
  tauri_build::try_build(attributes).expect("failed to run tauri-build")
}
