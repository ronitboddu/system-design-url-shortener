# High-Level Architecture

This document describes the current system architecture as implemented in the repository today.

## System Overview

The URL shortener is built as a small distributed system with three main runtime components:

- a public Go API gateway
- an internal Python DB service
- a PostgreSQL database

At a high level:

- the Go service handles public HTTP traffic
- the Python service owns persistence and ID generation
- PostgreSQL stores the durable URL records

This separation makes it easier to reason about responsibilities, scaling, and bottlenecks.

## Core Components

### 1. Go Service

The Go service is the public-facing entrypoint.

Its responsibilities are:

- accept shorten requests from clients
- accept redirect requests from clients
- call the Python DB service over HTTP
- build the final short URL returned to the client
- issue redirects for known short codes

The Go service runs on port `8080`.

### 2. Python DB Service

The Python service is an internal persistence layer implemented with FastAPI and SQLAlchemy.

Its responsibilities are:

- validate internal persistence requests
- deduplicate URLs based on `(original_url, ip_addr)`
- generate Snowflake IDs
- convert IDs into Base62 short codes
- read and write URL records in PostgreSQL

The Python service runs on port `8000`.

### 3. PostgreSQL

PostgreSQL is the durable storage layer.

It stores:

- original URLs
- short codes
- client IPs
- expiration values
- creation timestamps

## End-to-End Request Flow

The URL shortener currently follows this high-level request path:

- client
- Go service
- Python DB service
- PostgreSQL

For redirect reads, the client hits the Go service first, and the Go service forwards the lookup to the Python DB service, which reads from PostgreSQL.

For shorten requests, the client again hits the Go service first, and the Go service forwards the write request to the Python DB service, which handles deduplication, ID generation, and persistence.

## Write Path

The write path for `POST /shorten` is:

1. client sends a JSON request to the Go service
2. Go decodes the request and extracts the client IP
3. Go calls `POST /urls` on the Python DB service
4. Python checks whether the same `(original_url, ip_addr)` already exists
5. if a record exists, Python returns it
6. otherwise Python generates a Snowflake ID
7. Python converts that ID to a Base62 short code
8. Python stores the new record in PostgreSQL
9. Python returns the record to Go
10. Go builds the final `short_url` response for the client

## Read Path

The read path for `GET /{short_code}` is:

1. client requests a short code from the Go service
2. Go extracts the short code from the URL path
3. Go calls `GET /urls/{short_code}` on the Python DB service
4. Python queries PostgreSQL for the record
5. Python returns the original URL if found
6. Go redirects the client to the original URL

## Why the System Is Split Across Go and Python

This architecture is not the simplest possible design, but it is useful for learning and experimentation because it makes service boundaries explicit.

Benefits of this split:

- the public API tier is separate from the persistence tier
- service-to-service latency becomes visible under load
- scaling Go and Python independently becomes possible
- infrastructure experiments such as load balancing, ingress, and Kubernetes services become more meaningful

Tradeoff:

- every request currently depends on a synchronous network hop between Go and Python
- this adds latency compared to a single-process design

## Internal Configuration Model

### Go service configuration

The Go service reads:

- `DB_SERVICE_BASE_URL`

This tells the Go service where the Python DB service lives.

### Python DB service configuration

The Python service reads:

- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_SCHEMA`
- `SNOWFLAKE_NODE_ID`

This allows the same codebase to run:

- locally
- in Docker
- in Kubernetes

with different runtime wiring.

## Kubernetes Networking Roles

Once this system is deployed to Kubernetes, it helps to separate three different responsibilities:

- internal service-to-service routing
- internal load balancing across pods
- external traffic entry into the cluster

### 1. Kubernetes `Service`

A Kubernetes `Service` gives a stable name to a group of matching pods and distributes traffic across them.

For example:

- `go-service` selects all Go API pods
- `db-service` selects all Python DB-service pods

This means Kubernetes already provides internal pod-level load balancing.

So if there are multiple Go pods:

- traffic sent to `go-service`
- is automatically distributed across those Go pods

The same idea applies to the Python DB service.

### 2. Ingress

Ingress is used for external HTTP/HTTPS access into the cluster.

Its job is not to replace internal Services. Its job is to decide:

- which host should route to which service
- which path should route to which service

In this project, the Ingress should expose the Go service to external clients, while keeping the Python DB service internal.

That means the external flow becomes:

- client
- Ingress
- `go-service`
- Go pod

and the internal flow from Go to Python remains:

- Go pod
- `db-service`
- Python pod

### 3. Ingress Controller

The Ingress resource is only a routing rule. It does not handle traffic by itself.

An Ingress Controller, such as NGINX Ingress Controller, is the component that watches those rules and implements them.

Without an Ingress Controller:

- the Ingress YAML exists
- but external traffic will not actually be routed

## Important Design Distinction

Kubernetes `Service` objects already act like internal load balancers for pod-to-pod communication.

So inside the cluster:

- no additional custom NGINX or HAProxy layer is required just to spread traffic across Go replicas
- no additional custom load balancer is required just to spread traffic across Python replicas

However, for traffic coming from outside the cluster, you still need an entry mechanism such as:

- Ingress
- a `Service` of type `LoadBalancer`
- or, for local/dev setups, `NodePort` or `kubectl port-forward`

## Practical Lesson

In Docker, it made sense to run explicit NGINX load balancer containers in front of the Go and Python services.

In Kubernetes, that same pattern is usually unnecessary for internal traffic because:

- Kubernetes `Service` already handles internal load balancing

So the simpler Kubernetes architecture is:

- one Ingress for external HTTP entry
- one `Service` for Go pods
- one `Service` for Python DB-service pods
- PostgreSQL exposed internally by its own Service

This keeps the design closer to how production Kubernetes systems are typically structured.

## Current Architectural Characteristics

This system currently has a few important architectural properties:

- public traffic enters through the Go service
- persistence logic is centralized in the Python service
- PostgreSQL is the source of truth
- the redirect path is synchronous and database-dependent
- the write path includes deduplication before insert
- Kubernetes `Service` objects handle internal pod load balancing
- Ingress is used only for external HTTP entry

## Main Bottlenecks Observed So Far

From the load testing and debugging work, the main pressure points have been:

- synchronous Go -> Python -> PostgreSQL request chaining
- PostgreSQL connection pressure under load
- client disconnects and proxy-level failures at high concurrency
- network and routing mistakes when moving between Docker and Kubernetes

That makes this architecture a good learning platform because it exposes the real tradeoffs between:

- simplicity
- correctness
- horizontal scaling
- and operational complexity
