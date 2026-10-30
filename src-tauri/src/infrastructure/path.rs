use std::fs;
use std::io;
use std::path::{Path, PathBuf};

#[derive(Clone, Debug)]
pub struct DevboxLayout {
  root: PathBuf,
}

impl DevboxLayout {
  pub fn new(root: PathBuf) -> Self {
    Self { root }
  }

  pub fn managed_tool_root(&self, tool_id: &str) -> PathBuf {
    self.root.join("tools").join(tool_id)
  }

  pub fn active_installation_pointer(&self, tool_id: &str) -> PathBuf {
    self
      .root
      .join("state")
      .join("active")
      .join(format!("{tool_id}.json"))
  }

  pub fn config_file(&self) -> PathBuf {
    self.root.join("config.json")
  }

  pub fn config_backup_file(&self) -> PathBuf {
    self.root.join("config.json.bak")
  }
}

pub fn devbox_root(home_dir: &Path) -> PathBuf {
  home_dir.join(".devbox")
}

pub fn ensure_layout(root: &Path) -> io::Result<()> {
  let relative_directories = [
    "bin",
    "tools",
    "state",
    "state/active",
    "state/operations",
    "state/locks",
    "cache",
    "cache/releases",
    "cache/downloads",
    "staging",
    "backups",
    "logs",
  ];

  ensure_private_directory(root)?;
  for relative in relative_directories {
    ensure_private_directory(&root.join(relative))?;
  }

  Ok(())
}

fn ensure_private_directory(path: &Path) -> io::Result<()> {
  match fs::symlink_metadata(path) {
    Ok(metadata) if metadata.file_type().is_symlink() => {
      return Err(io::Error::new(
        io::ErrorKind::AlreadyExists,
        format!("拒绝使用符号链接目录: {}", path.display()),
      ));
    }
    Ok(metadata) if !metadata.is_dir() => {
      return Err(io::Error::new(
        io::ErrorKind::AlreadyExists,
        format!("路径已存在且不是目录: {}", path.display()),
      ));
    }
    Ok(_) => {}
    Err(error) if error.kind() == io::ErrorKind::NotFound => {
      fs::create_dir(path)?;
    }
    Err(error) => return Err(error),
  }

  restrict_permissions(path)
}

#[cfg(unix)]
fn restrict_permissions(path: &Path) -> io::Result<()> {
  use std::os::unix::fs::PermissionsExt;

  let mut permissions = fs::metadata(path)?.permissions();
  permissions.set_mode(0o700);
  fs::set_permissions(path, permissions)
}

#[cfg(not(unix))]
fn restrict_permissions(_path: &Path) -> io::Result<()> {
  Ok(())
}
