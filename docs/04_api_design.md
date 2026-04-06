# API Design

This document captures the API surface of the current URL shortener implementation as it exists in the Go gateway and Python DB service.

## API Layers

The system has two API layers:

- a public-facing Go API
- an internal Python DB-service API

The client is expected to talk to the Go API. The Go API then talks to the Python DB service over HTTP.

## Public Go API

The Go service runs on port `8080` and exposes two routes:

- `POST /shorten`
- `GET /{short_code}`

These routes are registered in [`server/cmd/api/main.go`](/Users/ronitboddu/Documents/Projects/system-design-url-shortener/server/cmd/api/main.go).

### 1. Create Short URL

**Endpoint**

```http
POST /shorten
```

**Purpose**

Accepts a long URL and returns a short URL that points back to the Go service.

**Expected request body**

```json
{
  "expTime": 2,
  "urlPath": "https://example.com/test"
}
```

**Behavior**

- the Go handler accepts only `POST`
- it decodes the request body into the service layer
- it extracts the client IP
- it forwards a normalized payload to the Python DB service
- it returns a JSON response containing the final short URL

**Successful response shape**

```json
{
  "short_url": "http://localhost:8080/abc123"
}
```

**Important implementation detail**

The short URL is currently constructed in the Go handler using:

- `http://localhost:8080/` + `short_code`

That means this is currently environment-specific and not yet configurable for production hostnames.

### 2. Redirect Lookup

**Endpoint**

```http
GET /{short_code}
```

**Purpose**

Looks up the original URL for a short code and redirects the client.

**Behavior**

- the handler extracts the short code from the path
- it asks the Python DB service for the record
- if found, it redirects to the original URL
- if not found, it returns `404`

This route is registered as:

```text
/
```

so it acts as a catch-all path for short-code lookups.

## Internal Python DB-Service API

The Python FastAPI service runs on port `8000` and exposes two internal endpoints:

- `POST /urls`
- `GET /urls/{short_code}`

These routes are defined in [`db-service/app/api/routes.py`](/Users/ronitboddu/Documents/Projects/system-design-url-shortener/db-service/app/api/routes.py).

### 1. Create or Reuse URL Record

**Endpoint**

```http
POST /urls
```

**Purpose**

Creates a new URL record, or returns the existing record if the same `(original_url, ip_addr)` pair already exists.

**Request shape**

```json
{
  "original_url": "https://example.com/test",
  "ip_addr": "10.0.0.1",
  "exp_time": 2
}
```

**Behavior**

- validates the request with Pydantic
- checks whether the same `original_url` and `ip_addr` already exist
- if yes, returns the existing short code
- if no, generates a new Snowflake ID
- converts the ID to a Base62 short code
- stores the record in PostgreSQL

**Response shape**

```json
{
  "original_url": "https://example.com/test",
  "short_code": "abc123",
  "exp_time": 2
}
```

### 2. Get URL Record

**Endpoint**

```http
GET /urls/{short_code}
```

**Purpose**

Fetches the stored record for a short code.

**Behavior**

- queries PostgreSQL by `short_code`
- returns `404` if the record does not exist
- otherwise returns the original URL and metadata

**Response shape**

```json
{
  "original_url": "https://example.com/test",
  "short_code": "abc123",
  "exp_time": 2
}
```

## Service-to-Service Contract

The Go service uses the Python service as its persistence layer.

This coupling is implemented in [`server/internal/client/db_service.go`](/Users/ronitboddu/Documents/Projects/system-design-url-shortener/server/internal/client/db_service.go).

### Go -> Python create flow

Go sends:

```json
{
  "original_url": "...",
  "ip_addr": "...",
  "exp_time": 2
}
```

to:

```http
POST /urls
```

### Go -> Python read flow

Go sends:

```http
GET /urls/{short_code}
```

and expects either:

- a URL record
- or `404`

## Error Handling

### Public Go API

- `POST /shorten`
  - returns `405` if the method is not `POST`
  - currently returns `404` if the downstream DB-service call fails
- `GET /{short_code}`
  - returns `404` if the short code is not found

### Python DB service

- `GET /urls/{short_code}`
  - returns `404` when the record does not exist
- `POST /urls`
  - currently relies on normal FastAPI / database exceptions for failure paths

## Current Design Characteristics

The current API design is intentionally simple:

- public traffic goes through one Go gateway
- persistence is isolated behind a Python service
- read and write APIs are synchronous
- deduplication is based on `original_url + ip_addr`

This makes the system easy to reason about, but it also means every request currently depends on a synchronous cross-service hop.
