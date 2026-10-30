use http::Uri;
use hyper_util::client::proxy::matcher::{Intercept, Matcher};

const HTTP_PROBE: &str = "http://devbox-system-proxy.invalid/";
const HTTPS_PROBE: &str = "https://devbox-system-proxy.invalid/";

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum SystemProxySource {
  Environment,
  OperatingSystem,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SystemProxyEndpoint {
  pub url: String,
  pub source: SystemProxySource,
  pub requires_authentication: bool,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SystemProxyInspection {
  pub http: Option<SystemProxyEndpoint>,
  pub https: Option<SystemProxyEndpoint>,
}

pub fn inspect_system_proxy() -> SystemProxyInspection {
  let environment = Matcher::from_env();
  let system = Matcher::from_system();

  SystemProxyInspection {
    http: resolve(&system, &environment, &Uri::from_static(HTTP_PROBE)),
    https: resolve(&system, &environment, &Uri::from_static(HTTPS_PROBE)),
  }
}

fn resolve(system: &Matcher, environment: &Matcher, target: &Uri) -> Option<SystemProxyEndpoint> {
  let matched = system.intercept(target)?;
  // `Matcher::from_system` starts with environment routes, then fills missing
  // HTTP/HTTPS routes from the OS. Comparing the complete selected route avoids
  // misattributing an ALL_PROXY fallback when the OS supplied the actual route.
  let source = if environment
    .intercept(target)
    .is_some_and(|environment| same_route(&environment, &matched))
  {
    SystemProxySource::Environment
  } else {
    SystemProxySource::OperatingSystem
  };

  Some(SystemProxyEndpoint {
    // hyper-util removes userinfo from the proxy URI and exposes it separately
    // as authentication data, so this value is safe to return to the frontend.
    url: sanitized_url(&matched),
    source,
    requires_authentication: requires_authentication(&matched),
  })
}

fn same_route(left: &Intercept, right: &Intercept) -> bool {
  left.uri() == right.uri()
    && left.basic_auth() == right.basic_auth()
    && left.raw_auth() == right.raw_auth()
}

fn sanitized_url(intercept: &Intercept) -> String {
  intercept.uri().to_string()
}

fn requires_authentication(intercept: &Intercept) -> bool {
  intercept.basic_auth().is_some() || intercept.raw_auth().is_some()
}
