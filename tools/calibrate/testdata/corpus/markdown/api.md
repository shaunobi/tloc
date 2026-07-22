# Batch export API

`POST /v1/exports` creates an asynchronous export. The request body accepts a
date range, output format, and optional list of fields.

```json
{
  "from": "2026-07-01T00:00:00Z",
  "to": "2026-08-01T00:00:00Z",
  "format": "csv",
  "fields": ["order_id", "created_at", "total"]
}
```

The response is `202 Accepted` with an export identifier and status URL. Polling
that URL returns `pending`, `running`, `complete`, or `failed`. Completed exports
include a signed download URL that expires after fifteen minutes.

Clients should send an `Idempotency-Key` header when creation may be retried.
Requests with the same key and normalized body return the original export.
Range limits are ninety days for CSV and thirty days for JSON Lines.
