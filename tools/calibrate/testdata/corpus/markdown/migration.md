# Migrating from configuration v2 to v3

Version 3 replaces the single `endpoint` field with a named `destinations`
array. Existing retry settings move under each destination, allowing independent
backoff policies.

```yaml
destinations:
  - name: primary
    url: https://events.example.test/ingest
    retry:
      attempts: 5
      maximum_delay: 30s
```

During rollout, the service accepts both schemas but refuses documents that mix
them. Run `config migrate --check` to preview the normalized output, then commit
the generated v3 document. Secrets remain references and are never copied into
the result.

The compatibility reader will be removed two minor releases after v3 becomes
the default. Dashboards expose a `configuration_schema_version` metric to find
remaining v2 deployments before that deadline.
