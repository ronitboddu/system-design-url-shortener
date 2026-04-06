# Database Design

This document captures the current PostgreSQL schema and the design decisions behind it.

## Database Role

PostgreSQL is the system of record for shortened URLs.

It stores:

- the original URL
- the generated short code
- the client IP used for deduplication
- the requested expiration time
- the creation timestamp

The Python DB service is the only component that talks to PostgreSQL directly.

## Schema

The application uses the PostgreSQL schema:

```text
tiny_url
```

The main table is:

```text
tiny_url.urls
```

This is defined in [`db-service/app/models/url.py`](/Users/ronitboddu/Documents/Projects/system-design-url-shortener/db-service/app/models/url.py).

## Table Structure

### Table: `tiny_url.urls`

| Column | Type | Purpose |
|---|---|---|
| `id` | `INTEGER` | primary key and Snowflake-derived unique identifier |
| `original_url` | `TEXT` | full destination URL |
| `short_code` | `VARCHAR(15)` | user-facing short code used in redirects |
| `ip_addr` | `VARCHAR(15)` | client IP used in deduplication logic |
| `exp_time` | `INTEGER` | expiration value provided by the client |
| `created_at` | `TIMESTAMP` | record creation time |

## Field Design Notes

### `id`

- generated in the Python service using a Snowflake-style ID generator
- acts as the primary key
- also serves as the input to Base62 encoding

### `short_code`

- derived by Base62-encoding the generated Snowflake ID
- exposed to the user in the shortened URL
- must be unique

### `original_url`

- stored as `TEXT` because URLs can vary in length
- this is the value returned on redirect lookup

### `ip_addr`

- currently used as part of the deduplication rule
- the same `original_url + ip_addr` pair should map to the same stored record

### `exp_time`

- currently stored as an integer exactly as provided by the API
- useful for future expiration enforcement
- not yet used to actively delete or invalidate records at read time

### `created_at`

- populated by PostgreSQL using `func.now()`
- useful for auditing and future cleanup jobs

## Constraints and Indexes

The table currently has three important index/constraint properties.

### 1. Primary key on `id`

This provides:

- row identity
- automatic uniqueness
- index support for primary-key access

### 2. Unique constraint on `short_code`

This ensures:

- no two records can share the same public short code
- redirect lookups by short code are efficient

PostgreSQL backs this with a unique B-tree index.

### 3. Composite unique index on `(original_url, ip_addr)`

The model defines:

- a unique index on `original_url` and `ip_addr`

This matches the repository behavior:

- before inserting, the DB service looks for an existing row with the same `(original_url, ip_addr)`
- if found, it returns the existing record instead of creating a new one

Why this matters:

- speeds up deduplication lookups
- enforces the deduplication rule at the database level
- protects against duplicate inserts under concurrency

## Query Patterns

The current system relies on two dominant database access patterns.

### 1. Write path deduplication

The Python repository does:

```sql
WHERE original_url = ? AND ip_addr = ?
```

This is why the composite unique index is important.

### 2. Redirect lookup

The Python repository does:

```sql
WHERE short_code = ?
```

This is why the unique index on `short_code` is essential.

## Data Access Layer

The repository logic lives in [`db-service/app/repositories/url_repository.py`](/Users/ronitboddu/Documents/Projects/system-design-url-shortener/db-service/app/repositories/url_repository.py).

It currently follows this flow:

### Insert flow

- open SQLAlchemy session
- start transaction
- query for existing `(original_url, ip_addr)`
- return existing row if found
- otherwise generate Snowflake ID
- encode Snowflake ID to Base62 `short_code`
- insert new row
- return plain dictionary data

### Read flow

- open SQLAlchemy session
- start transaction
- query by `short_code`
- return plain dictionary data or `None`

## Why Plain Dictionaries Are Returned

Earlier versions returned ORM objects directly, which caused session-lifetime problems such as detached instance errors.

The current design returns plain dictionaries from the repository so that:

- route handlers do not depend on a live SQLAlchemy session
- API responses are easier to build safely

## Current Tradeoffs

This schema is simple and well aligned with the current feature set, but it has a few known limitations:

- reads and writes both depend on synchronous PostgreSQL access
- `ip_addr` is part of correctness today, which may or may not be the final product rule
- `exp_time` is stored but not fully enforced yet
- the table is optimized for current lookup patterns, not yet for sharding or read replicas

## Why This Design Works Well for the Current System

The design is a good fit for the current architecture because:

- it supports deterministic short-code generation from Snowflake IDs
- it supports fast redirect lookup through `short_code`
- it supports duplicate suppression through `(original_url, ip_addr)`
- it keeps the persistence model simple while the service architecture is still evolving
