// Package postgresql provides a PostgreSQL database adapter for toutago-datamapper.
//
// This adapter implements the datamapper Adapter interface for PostgreSQL databases,
// with PostgreSQL-specific optimizations including RETURNING clauses for auto-generated IDs.
//
// # Features
//
//   - Full CRUD operations (Create, Read, Update, Delete)
//   - Bulk insert support with multi-row VALUES
//   - RETURNING clause for auto-generated IDs (SERIAL, IDENTITY)
//   - Named parameter substitution ({param_name})
//   - PostgreSQL-specific optimizations
//   - Optimistic locking support
//   - Connection pooling configuration
//   - Custom SQL execution and stored procedures
//   - CQRS pattern support via source configuration
//
// # Quick Start
//
// Register the PostgreSQL adapter with your datamapper:
//
//	import (
//	    "github.com/toutaio/toutago-datamapper/engine"
//	    "github.com/toutaio/toutago-datamapper-postgres"
//	)
//
//	mapper, err := engine.NewMapper("config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mapper.Close()
//
//	// Register PostgreSQL adapter
//	mapper.RegisterAdapter("postgresql", func(source config.Source) (adapter.Adapter, error) {
//	    return postgres.NewPostgreSQLAdapter(source.Connection)
//	})
//
// # Configuration
//
// Define PostgreSQL connection in YAML:
//
//	sources:
//	  - name: users_db
//	    type: postgresql
//	    config:
//	      host: localhost
//	      port: 5432
//	      user: myapp
//	      password: ${POSTGRES_PASSWORD}
//	      database: myapp_db
//	      sslmode: disable
//	      max_connections: 20
//
// # Usage
//
// Use through datamapper API:
//
//	user := &User{Name: "John", Email: "john@example.com"}
//	err := mapper.Insert(context.Background(), "User", user)
//	// user.ID is populated via RETURNING clause
//
//	err = mapper.Update(context.Background(), "User", user)
//
// # PostgreSQL-Specific Features
//
// RETURNING clause for auto-generated IDs:
//
//	INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id
//
// Connection pooling with PostgreSQL best practices:
//   - max_connections: Maximum open connections
//   - max_idle: Maximum idle connections
//   - sslmode: SSL connection mode
//
// # Thread Safety
//
// The adapter uses database/sql with lib/pq driver which provides
// connection pooling and is safe for concurrent use.
//
// # Version
//
// This is version 0.1.0 - requires toutago-datamapper v0.1.0 or higher.
// Requires Go 1.22 or higher.
package postgresql
