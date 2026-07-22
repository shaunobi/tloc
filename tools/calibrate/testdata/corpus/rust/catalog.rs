use std::collections::HashMap;
use std::fmt;
use std::str::FromStr;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Tier {
    Free,
    Standard,
    Enterprise,
}

#[derive(Debug, PartialEq, Eq)]
pub struct ParseTierError(String);

impl fmt::Display for ParseTierError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(formatter, "unknown tier: {}", self.0)
    }
}

impl std::error::Error for ParseTierError {}

impl FromStr for Tier {
    type Err = ParseTierError;

    fn from_str(value: &str) -> Result<Self, Self::Err> {
        match value.trim().to_ascii_lowercase().as_str() {
            "free" => Ok(Self::Free),
            "standard" => Ok(Self::Standard),
            "enterprise" => Ok(Self::Enterprise),
            other => Err(ParseTierError(other.to_owned())),
        }
    }
}

pub struct Catalog {
    limits: HashMap<String, u32>,
}

impl Catalog {
    pub fn from_pairs<'a>(pairs: impl IntoIterator<Item = (&'a str, u32)>) -> Self {
        let limits = pairs
            .into_iter()
            .map(|(name, limit)| (name.to_ascii_lowercase(), limit))
            .collect();
        Self { limits }
    }

    pub fn allowance(&self, feature: &str, tier: Tier) -> Option<u32> {
        let base = *self.limits.get(&feature.to_ascii_lowercase())?;
        Some(match tier {
            Tier::Free => base.min(10),
            Tier::Standard => base,
            Tier::Enterprise => base.saturating_mul(5),
        })
    }
}
