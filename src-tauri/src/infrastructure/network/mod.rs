mod client;

use std::error::Error;
use std::fmt;

pub use client::build_client;

#[derive(Clone, Debug)]
pub struct ProxyEndpoint {
  pub url: String,
  pub bypass: Vec<String>,
}

#[derive(Clone, Debug, Default)]
pub struct NetworkRoute {
  proxy: Option<ProxyEndpoint>,
}

impl NetworkRoute {
  pub fn direct() -> Self {
    Self { proxy: None }
  }

  pub fn manual(proxy: ProxyEndpoint) -> Self {
    Self { proxy: Some(proxy) }
  }

  pub fn proxy(&self) -> Option<&ProxyEndpoint> {
    self.proxy.as_ref()
  }
}

#[derive(Debug)]
pub enum NetworkClientError {
  InvalidProxy { message: String },
  Build { message: String },
  Request { message: String },
  Timeout,
  ProxyAuthenticationRequired,
}

impl fmt::Display for NetworkClientError {
  fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
    match self {
      Self::InvalidProxy { message } => write!(formatter, "代理配置无效: {message}"),
      Self::Build { message } => write!(formatter, "无法创建网络客户端: {message}"),
      Self::Request { message } => write!(formatter, "网络请求失败: {message}"),
      Self::Timeout => formatter.write_str("网络请求超时"),
      Self::ProxyAuthenticationRequired => formatter.write_str("代理要求身份验证"),
    }
  }
}

impl Error for NetworkClientError {}
