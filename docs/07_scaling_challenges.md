# Scaling Challenges

## Idle Transactions Under Load

During load testing, the Python DB service started showing a growing number of PostgreSQL connections in the `idle in transaction` state. At the same time, the `k6` threshold on `http_req_duration` was crossed, which showed that latency was rising under higher concurrency.

### Issue

The service was leaving database transactions open after request work had already finished. This was visible in `pg_stat_activity`, where multiple sessions remained in `idle in transaction` and some of them were stuck after the deduplication query:

```sql
SELECT ...
FROM tiny_url.urls
WHERE tiny_url.urls.original_url = $1
  AND tiny_url.urls.ip_addr = $2
LIMIT $3
```

This behavior is dangerous at scale because open idle transactions can:

- hold locks longer than necessary
- increase connection pressure
- slow down concurrent requests
- increase tail latency during load tests

### Root Cause

The main problem was transaction/session handling in the Python repository layer.

- The deduplication lookup in `put_record()` opened a transaction and did not always end it cleanly.
- The repository method mixed manual session control with query logic.
- `session.refresh(record)` added extra session-state complexity and produced `InvalidRequestError` in some cases because the ORM instance was no longer persistent in the expected session state.

As a result, some requests completed their query work, but PostgreSQL still saw those connections as being inside open transactions.

### Solutions Tried

Several checks were used to isolate the issue:

- `pg_stat_activity` was queried repeatedly to track connection states.
- `SELECT pid, state, xact_start, query_start, query FROM pg_stat_activity WHERE state = 'idle in transaction';` was used to identify the exact SQL statements leaving transactions open.
- The repository query logic was reviewed to confirm that the deduplication path was the common source of stuck transactions.

### What Worked

The fix was to simplify transaction management in the Python DB service:

- use a managed transaction block with `with self.session_factory.begin() as session:`
- keep the deduplication query and insert logic inside that single managed block
- remove unnecessary manual transaction handling
- remove `session.refresh(record)` from the write path when it was not needed for the API response

After these changes:

- `idle in transaction` connections no longer appeared during testing
- transaction cleanup became automatic on success and rollback on failure
- PostgreSQL connection state became healthier under load

### Takeaway

At higher request volumes, even small mistakes in ORM session handling can become a real scaling bottleneck. Explicitly using managed transaction boundaries in the repository layer made the system more predictable and reduced database-side contention.
