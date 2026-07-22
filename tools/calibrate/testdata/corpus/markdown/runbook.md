# Runbook: elevated checkout failures

Use this procedure when the five-minute checkout error rate exceeds three
percent. Start an incident channel and record every change in the timeline.

## Triage

- Compare failures by region, payment provider, and application version.
- Check the most recent deployment and feature-flag audit entries.
- Inspect dependency latency before increasing worker capacity.
- Confirm that synthetic checkout probes fail in the same way.

If one provider is unhealthy, route new sessions to the fallback and preserve
the original provider reference for reconciliation. If a release introduced the
regression, pause rollout and trigger the automated rollback. Do not retry
ambiguous payment authorizations until their status has been queried.

Resolution requires thirty quiet minutes, a reconciled list of affected orders,
and a short customer-impact summary linked from the incident record.
