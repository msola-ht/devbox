use std::time::Duration;

use reqwest::{Client, NoProxy, Proxy, redirect::Policy};

use super::{NetworkClientError, NetworkRoute};

const CONNECT_TIMEOUT: Duration = Duration::from_secs(10);
const MAX_REDIRECTS: usize = 5;

pub fn build_client(route: &NetworkRoute) -> Result<Client, NetworkClientError> {
  let mut builder = Client::builder()
    .https_only(true)
    .connect_timeout(CONNECT_TIMEOUT)
    .redirect(Policy::limited(MAX_REDIRECTS))
    .user_agent(format!("DevBox/{}", env!("CARGO_PKG_VERSION")));

  match route {
    NetworkRoute::System => {}
    NetworkRoute::Direct => {
      builder = builder.no_proxy();
    }
    NetworkRoute::Manual(endpoint) => {
      let proxy = Proxy::all(&endpoint.url).map_err(|error| NetworkClientError::InvalidProxy {
        message: error.to_string(),
      })?;
      let bypass = if endpoint.bypass.is_empty() {
        None
      } else {
        NoProxy::from_string(&endpoint.bypass.join(","))
      };
      builder = builder.no_proxy().proxy(proxy.no_proxy(bypass));
    }
  }

  builder.build().map_err(|error| NetworkClientError::Build {
    message: error.to_string(),
  })
}
