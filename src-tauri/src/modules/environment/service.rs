use super::EnvironmentInfo;

#[derive(Clone, Debug)]
pub struct EnvironmentService {
  info: EnvironmentInfo,
}

impl EnvironmentService {
  pub fn new(info: EnvironmentInfo) -> Self {
    Self { info }
  }

  pub fn info(&self) -> &EnvironmentInfo {
    &self.info
  }
}
