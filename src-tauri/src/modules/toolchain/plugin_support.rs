use std::path::{Path, PathBuf};

pub(super) fn parse_version_line(output: &str) -> Option<String> {
  output
    .lines()
    .find_map(|line| line.split_whitespace().find_map(parse_version_token))
}

pub(super) fn parse_version_token(token: &str) -> Option<String> {
  let version = token.strip_prefix('v').unwrap_or(token);
  let mut characters = version.chars();
  if !characters.next()?.is_ascii_digit()
    || !version.contains('.')
    || !characters
      .all(|character| character.is_ascii_alphanumeric() || matches!(character, '.' | '-' | '+'))
  {
    return None;
  }
  Some(version.to_string())
}

pub(super) fn normalized_path(path: &Path) -> String {
  path.to_string_lossy().replace('\\', "/").to_lowercase()
}

pub(super) fn default_shell_candidate(
  tool_id: &str,
  default_shell: Option<&str>,
) -> Option<PathBuf> {
  let shell = PathBuf::from(default_shell?);
  if !shell.is_absolute()
    || shell.file_stem().and_then(|name| name.to_str()) != Some(tool_id)
    || !shell.is_file()
  {
    return None;
  }
  Some(shell)
}
