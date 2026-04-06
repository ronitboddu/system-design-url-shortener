# Engineering Retrospective

This document captures the main implementation and debugging challenges faced while building the URL shortener, along with the reasoning process used to isolate and resolve them.

## 1. Idle Transactions in PostgreSQL

### Problem

Under load, PostgreSQL showed multiple connections in the `idle in transaction` state. This caused concern because open transactions can hold locks, consume connections, and contribute to rising tail latency.

### How It Was Investigated

- Repeatedly queried `pg_stat_activity` to count connection states.
- Inspected only `idle in transaction` sessions:

```sql
SELECT pid, state, xact_start, query_start, query
FROM pg_stat_activity
WHERE state = 'idle in transaction';
```

- Observed that the most common stuck query was the deduplication read in the Python repository:

```sql
SELECT ...
FROM tiny_url.urls
WHERE original_url = $1 AND ip_addr = $2
LIMIT $3;
```

### Root Cause

The Python repository layer was mixing ORM session usage with transaction handling in a way that left transactions open longer than expected.

### What Was Tried

- Reviewed the read and write repository methods.
- Simplified the transaction structure.
- Removed unnecessary ORM refresh logic.

### What Worked

- Managed the transaction inside the session more explicitly.
- Kept the deduplication read and insert logic within a cleaner transaction boundary.
- Removed `session.refresh(record)` when it was not needed.

### Lesson

Even a simple ORM-backed repository can create scaling issues if session and transaction boundaries are not carefully controlled.

## 2. ORM Object Lifetime / Detached Instance Problems

### Problem

Several errors appeared while returning SQLAlchemy ORM objects from the repository layer:

- `DetachedInstanceError`
- `InvalidRequestError`
- object not persistent within session

### Thought Process

The failures only appeared after route handlers tried to access fields like:

- `record.original_url`
- `record.short_code`

This pointed to a timing issue between session closure and response construction.

### Root Cause

Repository methods returned ORM objects after the session/transaction had ended, and later attribute access triggered SQLAlchemy to refresh expired state.

### What Worked

- Converted repository return values into plain dictionaries before leaving the session scope.
- Made route handlers build API responses from those plain values instead of directly from ORM objects.

### Lesson

Repository boundaries are easier to reason about when they return plain data structures rather than live ORM-managed objects.

## 3. Realistic Load-Test Latency

### Problem

The initial system looked healthy for individual requests, but once mixed read/write traffic was increased, `k6` showed severe degradation:

- `p(95)` above multiple seconds
- `p(99)` above multiple seconds

### Thought Process

The first question was whether this was:

- a broken implementation
- a database issue
- or a real architectural bottleneck

By combining:

- `k6` percentile metrics
- Go-side timing logs
- Python route timing logs
- PostgreSQL activity inspection

it became clear that the problem was not a single broken query. It was cumulative synchronous latency across multiple service hops.

### Main Insight

The request path was:

- client / `k6`
- Go service
- Python DB service
- PostgreSQL

Even if each hop looked acceptable in isolation, queueing under higher concurrency created visible end-to-end latency.

### Lesson

End-to-end latency in distributed systems often emerges from the combination of multiple acceptable local costs.

## 4. Connection and Proxy-Level Failures Under Load

### Problem

During heavier load tests, requests started failing with:

- `EOF`
- `connection reset by peer`
- Nginx `499` responses

### Thought Process

At first this looked like:

- service crashes
- broken upstream routing
- or bad container networking

But inspection showed:

- Python load balancer logs still had successful `200` responses
- Nginx on the Go front door showed `499`

### Root Cause

`499` in Nginx means the client closed the connection before the response completed. This indicated front-end queueing / overload rather than a pure service-to-service connectivity failure.

### Lesson

Not all failed requests are application errors. Some are symptoms of overload where the client gives up before the backend finishes.

## 5. Docker Networking Misconfiguration

### Problem

After containerizing the services, requests were reaching Nginx but not behaving as expected. Some requests returned `404`, and data written through the Python service did not appear in the expected PostgreSQL container.

### Thought Process

Each hop had to be validated separately:

- `k6` -> Go load balancer
- Go load balancer -> Go services
- Go services -> Python load balancer
- Python load balancer -> Python services
- Python services -> PostgreSQL

### Root Causes

Several configuration issues were discovered:

#### a. Wrong upstream name in Go service config

The Go containers were configured with:

```text
DB_SERVICE_BASE_URL=http://db-lb:8000
```

but the actual Python load balancer container was named `python-lb`.

#### b. Python service connected to host Postgres instead of container Postgres

The Python container environment showed:

```text
DB_HOST=host.docker.internal
```

while debugging was being done against the `postgres-db` container.

#### c. Nginx upstream confusion

Nginx startup logs helped reveal when upstream hostnames could not be resolved or when requests were not being proxied to the expected backend.

### What Worked

- Verified container names using `docker ps`.
- Verified Docker network membership.
- Verified live environment variables inside running containers with `docker exec`.
- Checked mounted Nginx configs directly inside the containers.

### Lesson

Containerized systems introduce a new class of debugging problems where application logic may be correct, but the deployed topology is wrong.

## 6. Snowflake IDs in a Multi-Replica Python Service

### Problem

Once the Python DB service was replicated, Snowflake-based ID generation could no longer assume a single process.

### Concern

If multiple Python service replicas share the same Snowflake node ID, duplicate Snowflake IDs can be generated.

### Resolution Direction

- Introduced `SNOWFLAKE_NODE_ID` as an environment-driven setting.
- Planned unique node IDs per DB-service replica.

### Lesson

Algorithms that are safe in a single process can become unsafe immediately when horizontal scaling is introduced.

## 7. Current Scaling Direction

After correctness and infrastructure issues were stabilized, the next improvements identified were:

- indexing the most common lookup paths
- sharding only if a single database becomes the limiting factor
- caching redirect lookups to reduce repeated database reads

These were chosen because the traffic profile is read-heavy and because the biggest remaining scaling opportunity is reducing synchronous database dependence on the redirect path.

## 8. Kubernetes Ingress Debugging in a Local Minikube Setup

### Problem

After deploying the Go service behind an NGINX Ingress in Kubernetes, requests to:

- `http://shortener.local/shorten`

did not work reliably. Depending on the setup stage, the failures looked like:

- `connection refused`
- request hangs / timeouts
- `404` from the ingress path

This was confusing because the Ingress resource itself appeared to be created correctly.

### Thought Process

The debugging process had to separate three different concerns:

- whether DNS name resolution was correct
- whether the local machine could actually reach the Ingress controller
- whether the Ingress rule matched the incoming request host

It was important not to treat all of these as the same problem, because each one lives at a different layer.

### What Was Verified

- `kubectl get ingress -n url-shortener` showed:
  - host `shortener.local`
  - ingress class `nginx`
  - backend `go-service:8080`
- `kubectl describe ingress ...` showed that the rule had been accepted and synced by the ingress controller.
- `kubectl get pods -A | grep ingress` confirmed the `ingress-nginx-controller` pod was running.
- `kubectl get svc -A | grep ingress` showed that the ingress controller service was `NodePort`, not `LoadBalancer`.

### Why the Initial Attempts Failed

#### a. `shortener.local -> 127.0.0.1` without a local listener

At one point `/etc/hosts` mapped:

```text
127.0.0.1 shortener.local
```

but nothing was actually listening on local port `80`, so:

- DNS worked
- the HTTP connection still failed

#### b. `shortener.local -> minikube IP` without host reachability

The mapping was then changed to the Minikube IP:

```text
192.168.49.2 shortener.local
```

However, the Mac host could not directly reach that Minikube IP. Both:

- `ping $(minikube ip)`
- `curl http://shortener.local/...`

timed out, which showed that the Ingress was configured, but the host-to-cluster network path was not usable.

#### c. `minikube tunnel` was not the right fix for this specific service shape

`minikube tunnel` was tried in order to avoid repeated `kubectl port-forward` usage. However, the ingress controller service was of type `NodePort`, not `LoadBalancer`.

That meant:

- the tunnel process could start
- but it did not produce a working listener on local port `80`

This was confirmed by:

```bash
lsof -iTCP:80 -sTCP:LISTEN
```

which showed that nothing was actually bound on the host.

### What Finally Worked

The working solution was:

```bash
kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80
```

and then sending requests to:

```text
http://127.0.0.1:8080
```

with the correct host header:

```text
Host: shortener.local
```

For example:

```bash
curl -H "Host: shortener.local" http://127.0.0.1:8080/shorten
```

This returned:

- `method not allowed`

which was actually a good sign, because it proved:

- the request reached the Ingress controller
- the Ingress rule matched `shortener.local`
- the request was forwarded to the Go service
- the Go route existed but expected `POST`, not `GET`

### Why `k6` Was Returning `404`

The `k6` scripts were sending requests to:

- `http://localhost:8080/...`

without setting:

- `Host: shortener.local`

The Ingress rule was host-based, so requests with:

- `Host: localhost`
- or `Host: 127.0.0.1`

did not match the Ingress rule and produced `404`.

The fix was to add the host header explicitly in the `k6` request options:

```text
Host: shortener.local
```

### Lesson

Ingress debugging in local Kubernetes has at least three distinct layers:

- hostname resolution
- host reachability to the ingress entrypoint
- ingress rule matching based on host/path

In this case, the Ingress resource itself was correct. The failures came from:

- unreachable Minikube networking from the host
- using a `NodePort` ingress controller while expecting localhost port `80` behavior
- missing `Host` headers in local `curl`/`k6` tests

The final reliable local workflow was:

- port-forward the ingress controller service
- send traffic to `127.0.0.1:8080`
- include `Host: shortener.local` in requests

## Summary of Key Takeaways

- Transaction handling bugs can create database pressure even before query performance becomes the problem.
- Returning ORM objects across closed session boundaries makes debugging harder than returning plain data.
- End-to-end latency must be measured across service boundaries, not only inside one service.
- Nginx `499` and client disconnects are important overload signals.
- Docker and service-discovery mistakes can look like application bugs until each network hop is validated explicitly.
- Horizontal scaling often forces previously hidden assumptions, such as unique Snowflake node IDs, into the open.
- Local Kubernetes Ingress debugging often fails because of host networking and host-header mismatches, even when the Ingress YAML itself is correct.
