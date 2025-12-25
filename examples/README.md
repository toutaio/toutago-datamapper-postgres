# PostgreSQL Adapter Examples

Complete working examples demonstrating the PostgreSQL adapter.

## Prerequisites

- PostgreSQL database running
- Go 1.21+

## Quick Setup with Docker

```bash
docker run -d --name postgres-dev \
  -e POSTGRES_PASSWORD=devpass \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  postgres:16-alpine
```

## Examples

1. **Basic CRUD** - Simple create, read, update, delete
2. **Bulk Operations** - Efficient batch inserts

## Cleanup

```bash
docker rm -f postgres-dev
```
