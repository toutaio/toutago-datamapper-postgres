# PostgreSQL Adapter for Toutago DataMapper

A PostgreSQL database adapter implementation for [toutago-datamapper](https://github.com/toutaio/toutago-datamapper), providing full CRUD operations, bulk inserts, RETURNING clause for generated IDs, and custom query execution.

## Features

- ✅ Full CRUD operations (Create, Read, Update, Delete)
- ✅ Bulk insert support with multi-row VALUES
- ✅ RETURNING clause for auto-generated IDs (SERIAL, IDENTITY)
- ✅ Named parameter substitution (`{param_name}`)
- ✅ PostgreSQL-specific optimizations
- ✅ Optimistic locking support
- ✅ Connection pooling configuration
- ✅ Custom SQL execution and stored procedures
- ✅ CQRS pattern support via source configuration

## Installation

```bash
go get github.com/toutaio/toutago-datamapper-postgres
```

## Quick Start

### 1. Define Your Configuration

Create a `config.yaml` file with PostgreSQL connection details:

```yaml
sources:
  - name: users_db
    type: postgresql
    config:
      host: localhost
      port: 5432
      user: myapp
      password: ${POSTGRES_PASSWORD}
      database: myapp_db
      sslmode: disable
      max_connections: 20
      max_idle: 5
      conn_max_age_seconds: 3600

mappings_path: ./mappings
```

### 2. Configure Mappings

Create a `mappings/users.yaml` file:

```yaml
object: User
source: users_db

mappings:
  - name: fetch_by_id
    type: fetch
    statement: "SELECT id, name, email, created_at FROM users WHERE id = {id}"
    properties:
      - object: ID
        data: id
      - object: Name
        data: name
      - object: Email
        data: email
      - object: CreatedAt
        data: created_at

  - name: insert
    type: insert
    statement: users
    properties:
      - object: Name
        data: name
      - object: Email
        data: email
    generated:
      - object: ID
        data: id
```

### 3. Use in Your Application

```go
package main

import (
    "context"
    "log"
    
    "github.com/toutaio/toutago-datamapper/engine"
    "github.com/toutaio/toutago-datamapper/config"
    "github.com/toutaio/toutago-datamapper/adapter"
    postgresql "github.com/toutaio/toutago-datamapper-postgres"
)

func main() {
    mapper, err := engine.NewMapper("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    defer mapper.Close()
    
    mapper.RegisterAdapter("postgresql", func(source config.Source) (adapter.Adapter, error) {
        return postgresql.NewPostgreSQLAdapter(), nil
    })
    
    ctx := context.Background()
    
    user := map[string]interface{}{
        "Name":  "Alice Johnson",
        "Email": "alice@example.com",
    }
    
    if err := mapper.Insert(ctx, "users.insert", user); err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created user with ID: %v", user["ID"])
}
```

## PostgreSQL-Specific Features

### RETURNING Clause

PostgreSQL's RETURNING clause is automatically used for generated IDs:

```sql
INSERT INTO users (name) VALUES ($1) RETURNING id, created_at
```

### Bulk Inserts

Multi-row inserts are optimized:

```sql
INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4), ($5, $6)
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `host` | `localhost` | PostgreSQL server hostname |
| `port` | `5432` | PostgreSQL server port |
| `user` | `postgres` | Database user |
| `password` | - | Database password |
| `database` | - | Database name |
| `sslmode` | `disable` | SSL mode: disable, require, verify-ca, verify-full |
| `max_connections` | `10` | Maximum open connections |
| `max_idle` | `5` | Maximum idle connections |
| `conn_max_age_seconds` | `3600` | Connection max lifetime |

## Testing

```bash
go test -v
```

## Related Projects

- [toutago-datamapper](https://github.com/toutaio/toutago-datamapper) - Core library
- [toutago-datamapper-mysql](https://github.com/toutaio/toutago-datamapper-mysql) - MySQL adapter

## License

MIT License
