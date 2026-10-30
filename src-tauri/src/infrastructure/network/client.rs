use std::time::Duration;

use reqwest::{Client, NoProxy, Proxy, redirect::Policy};

use super::{NetworkClientError, NetworkRoute};

const CONNECT_TIMEOUT: Duration = Duration::from_secs(10);
const MAX_REDIRECTS: usize = 5;

pub fn build_client(route: &NetworkRoute) -> Result<Client, NetworkClientError> {
  let mut builder = Client::builder()
    .no_proxy()
    .https_only(true)
    .connect_timeout(CONNECT_TIMEOUT)
    .redirect(Policy::limited(MAX_REDIRECTS))
    .user_agent(format!("DevBox/{}", env!("CARGO_PKG_VERSION")));

  if let Some(endpoint) = route.proxy() {
    let proxy = Proxy::all(&endpoint.url).map_err(|error| NetworkClientError::InvalidProxy {
      message: error.to_string(),
    })?;
    let bypass = if endpoint.bypass.is_empty() {
      None
    } else {
      NoProxy::from_string(&endpoint.bypass.join(","))
    };
    builder = builder.proxy(proxy.no_proxy(bypass));
  }

  builder.build().map_err(|error| NetworkClientError::Build {
    message: error.to_string(),
  })
}
