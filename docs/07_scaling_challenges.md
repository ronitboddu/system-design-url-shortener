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

## High Read/Write Latency Under Sustained Load

After fixing the idle transaction issue, the system was pushed with a mixed `k6` workload using an 80:20 read:write ratio. By increasing the request rate, the application started showing real end-to-end latency degradation.

### Issue

Under higher mixed traffic, `k6` reported severe latency:

- `avg` latency rose to multiple seconds
- `p(95)` rose above 5 seconds
- `p(99)` also rose above 5 seconds

This meant the system was no longer just showing occasional spikes. Requests were now queueing under sustained load, which made this a realistic scaling bottleneck rather than a one-off implementation bug.

### Root Cause

The latency at this stage appears to come from architectural pressure rather than a single broken query. The request path contains several synchronous hops:

- `k6` sends traffic to the Go service
- the Go service synchronously calls the Python DB service
- the Python DB service synchronously queries PostgreSQL

Even when individual Go and Python execution logs looked small for sampled requests, the combined pipeline still accumulated queueing delay under load. This is consistent with a system that is being pushed beyond the capacity of one or more shared bottlenecks such as:

- PostgreSQL query/index performance
- Python DB service throughput
- Go to Python network hop overhead
- lack of a cache for read-heavy redirect traffic

### Observations

Several data points helped confirm that this was a true scaling issue:

- `k6` had to raise active VUs significantly to maintain the configured arrival rate
- Go service logs showed request duration spikes, especially on redirect traffic
- Python DB service logs showed request handling was usually fast in isolation, which suggests queueing and shared bottlenecks are important contributors

### Solutions Planned

The next scaling solutions to evaluate are:

#### 1. Indexing

The first step is to make sure the hottest lookups are indexed properly, especially on fields used in:

- redirect lookups by `short_code`
- deduplication lookups by `original_url` and `ip_addr`

The goal is to reduce query time and lower database CPU and I/O pressure before introducing more complex infrastructure.

#### 2. Sharding

If a single PostgreSQL instance becomes the long-term bottleneck, the next step is to shard the data across multiple database instances. This would distribute read and write pressure instead of forcing all traffic through one node.

This is a more advanced scaling step and adds operational complexity, so it is planned after indexing is evaluated.

#### 3. Caching

Because redirect traffic is much more frequent than write traffic, caching is a natural optimization. The likely cache path is:

- cache `short_code -> original_url`
- serve redirect reads from cache on hits
- fall back to the database on cache misses

This should reduce repeated database reads for hot short URLs and lower tail latency significantly for read-heavy workloads.

### Takeaway

Once correctness problems were fixed, the next bottleneck became overall system throughput under mixed traffic. At that point, the focus shifts from debugging code paths to classic scaling strategies: better indexing, distributing data, and introducing a cache for the read-heavy path.
