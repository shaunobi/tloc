use std::collections::BTreeMap;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Record {
    pub section: String,
    pub key: String,
    pub value: String,
}

pub fn parse(input: &str) -> Result<Vec<Record>, String> {
    let mut section = String::from("default");
    let mut records = Vec::new();
    for (line_number, raw) in input.lines().enumerate() {
        let line = raw.trim();
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        if let Some(name) = line.strip_prefix('[').and_then(|v| v.strip_suffix(']')) {
            section = name.trim().to_owned();
            continue;
        }
        let (key, value) = line
            .split_once('=')
            .ok_or_else(|| format!("line {} is missing '='", line_number + 1))?;
        records.push(Record {
            section: section.clone(),
            key: key.trim().to_owned(),
            value: value.trim().to_owned(),
        });
    }
    Ok(records)
}

pub fn index(records: &[Record]) -> BTreeMap<(&str, &str), &str> {
    records.iter().map(|r| ((r.section.as_str(), r.key.as_str()), r.value.as_str())).collect()
}
