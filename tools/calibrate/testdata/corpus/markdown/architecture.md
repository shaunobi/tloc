# Event delivery architecture

The delivery service accepts domain events, stores them durably, and fans them
out to subscribers without making the request path wait for remote systems.
Producers provide a stable event identifier so retries remain idempotent.

## Flow

1. Validate the envelope and normalize its timestamp to UTC.
2. Insert the payload and an outbox row in one database transaction.
3. Lease pending rows in bounded batches.
4. Publish each row, recording the broker acknowledgement before releasing it.

Workers use exponential backoff with jitter. A row moves to the dead-letter
queue after twelve attempts, where operators can inspect and replay it. Metrics
cover queue age, attempts, publish latency, and dead-letter volume.

The system favors at-least-once delivery. Consumers must therefore reject
duplicate identifiers or make their updates naturally idempotent.
