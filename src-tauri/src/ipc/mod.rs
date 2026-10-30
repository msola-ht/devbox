pub mod environment;
pub mod error;
pub mod toolchain;

use time::OffsetDateTime;
use time::format_description::well_known::Rfc3339;

pub(super) fn now_rfc3339() -> String {
  OffsetDateTime::now_utc()
    .format(&Rfc3339)
    .unwrap_or_else(|_| "1970-01-01T00:00:00Z".to_string())
}
