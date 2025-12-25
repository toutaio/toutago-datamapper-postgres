// Package postgresql provides a PostgreSQL adapter implementation for toutago-datamapper.
// This adapter enables mapping domain objects to PostgreSQL database tables with full CRUD support.
package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/toutaio/toutago-datamapper/adapter"
	_ "github.com/lib/pq"
)

// PostgreSQLAdapter implements the adapter.Adapter interface for PostgreSQL databases.
type PostgreSQLAdapter struct {
	db         *sql.DB
	dsn        string
	maxConn    int
	maxIdle    int
	connMaxAge int
}

// Config keys for PostgreSQL adapter configuration
const (
	ConfigHost     = "host"
	ConfigPort     = "port"
	ConfigUser     = "user"
	ConfigPassword = "password"
	ConfigDatabase = "database"
	ConfigSSLMode  = "sslmode"
	ConfigMaxConn  = "max_connections"
	ConfigMaxIdle  = "max_idle"
	ConfigConnAge  = "conn_max_age_seconds"
)

// NewPostgreSQLAdapter creates a new PostgreSQL adapter instance.
func NewPostgreSQLAdapter() *PostgreSQLAdapter {
	return &PostgreSQLAdapter{
		maxConn:    10,
		maxIdle:    5,
		connMaxAge: 3600,
	}
}

// Name returns the adapter type identifier.
func (a *PostgreSQLAdapter) Name() string {
	return "postgresql"
}

// Connect establishes connection to PostgreSQL database.
func (a *PostgreSQLAdapter) Connect(ctx context.Context, config map[string]interface{}) error {
	// Extract connection parameters
	host := getStringConfig(config, ConfigHost, "localhost")
	port := getIntConfig(config, ConfigPort, 5432)
	user := getStringConfig(config, ConfigUser, "postgres")
	password := getStringConfig(config, ConfigPassword, "")
	database := getStringConfig(config, ConfigDatabase, "")
	sslMode := getStringConfig(config, ConfigSSLMode, "disable")

	// Optional connection pooling parameters
	if maxConn, ok := config[ConfigMaxConn].(int); ok {
		a.maxConn = maxConn
	}
	if maxIdle, ok := config[ConfigMaxIdle].(int); ok {
		a.maxIdle = maxIdle
	}
	if connAge, ok := config[ConfigConnAge].(int); ok {
		a.connMaxAge = connAge
	}

	// Build DSN (connection string)
	a.dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslMode)

	// Open database connection
	db, err := sql.Open("postgres", a.dsn)
	if err != nil {
		return fmt.Errorf("postgresql: failed to open connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(a.maxConn)
	db.SetMaxIdleConns(a.maxIdle)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("postgresql: failed to ping database: %w", err)
	}

	a.db = db
	return nil
}

// Close releases database connections.
func (a *PostgreSQLAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Fetch retrieves one or more records from the database.
func (a *PostgreSQLAdapter) Fetch(ctx context.Context, op *adapter.Operation, params map[string]interface{}) ([]interface{}, error) {
	if a.db == nil {
		return nil, fmt.Errorf("postgresql: not connected")
	}

	query := op.Statement
	args, err := extractArgs(query, params)
	if err != nil {
		return nil, err
	}
	query = replaceNamedParams(query)

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgresql: query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("postgresql: failed to get columns: %w", err)
	}

	// Scan results
	var results []interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("postgresql: scan failed: %w", err)
		}

		// Build result map
		result := make(map[string]interface{})
		for i, col := range columns {
			result[col] = values[i]
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgresql: rows iteration failed: %w", err)
	}

	if len(results) == 0 && !op.Multi {
		return nil, adapter.ErrNotFound
	}

	return results, nil
}

// Insert creates new records in the database.
func (a *PostgreSQLAdapter) Insert(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	if a.db == nil {
		return fmt.Errorf("postgresql: not connected")
	}

	if len(objects) == 0 {
		return nil
	}

	// PostgreSQL supports RETURNING clause for generated IDs
	if len(op.Generated) > 0 {
		return a.insertWithReturning(ctx, op, objects)
	}

	return a.insertBulk(ctx, op, objects)
}

// insertWithReturning handles inserts with RETURNING clause for generated columns
func (a *PostgreSQLAdapter) insertWithReturning(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	tableName := op.Statement
	columns := make([]string, len(op.Properties))
	for i, prop := range op.Properties {
		columns[i] = prop.DataField
	}

	// Build RETURNING clause
	returningCols := make([]string, len(op.Generated))
	for i, gen := range op.Generated {
		returningCols[i] = gen.DataField
	}

	for _, objInterface := range objects {
		obj := objInterface.(map[string]interface{})
		placeholders := make([]string, len(columns))
		values := make([]interface{}, len(columns))
		for i, prop := range op.Properties {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			values[i] = obj[prop.ObjectField]
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
			strings.Join(returningCols, ", "))

		// Scan generated values
		scanDest := make([]interface{}, len(op.Generated))
		for i := range op.Generated {
			var val interface{}
			scanDest[i] = &val
		}

		if err := a.db.QueryRowContext(ctx, query, values...).Scan(scanDest...); err != nil {
			return fmt.Errorf("postgresql: insert with returning failed: %w", err)
		}

		// Set generated values back to object
		for i, gen := range op.Generated {
			val := *(scanDest[i].(*interface{}))
			obj[gen.ObjectField] = val
		}
	}

	return nil
}

// insertBulk handles bulk inserts without generated columns
func (a *PostgreSQLAdapter) insertBulk(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	tableName := op.Statement
	columns := make([]string, len(op.Properties))
	for i, prop := range op.Properties {
		columns[i] = prop.DataField
	}

	// Build multi-row insert
	valueRows := make([]string, len(objects))
	allValues := make([]interface{}, 0, len(objects)*len(columns))
	paramIndex := 1

	for i, objInterface := range objects {
		obj := objInterface.(map[string]interface{})
		placeholders := make([]string, len(columns))
		for j, prop := range op.Properties {
			placeholders[j] = fmt.Sprintf("$%d", paramIndex)
			paramIndex++
			allValues = append(allValues, obj[prop.ObjectField])
		}
		valueRows[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(valueRows, ", "))

	_, err := a.db.ExecContext(ctx, query, allValues...)
	if err != nil {
		return fmt.Errorf("postgresql: bulk insert failed: %w", err)
	}

	return nil
}

// Update modifies existing records in the database.
func (a *PostgreSQLAdapter) Update(ctx context.Context, op *adapter.Operation, objects []interface{}) error {
	if a.db == nil {
		return fmt.Errorf("postgresql: not connected")
	}

	query := op.Statement
	for _, objInterface := range objects {
		obj := objInterface.(map[string]interface{})
		args, err := extractArgs(query, obj)
		if err != nil {
			return err
		}
		pgQuery := replaceNamedParams(query)

		result, err := a.db.ExecContext(ctx, pgQuery, args...)
		if err != nil {
			return fmt.Errorf("postgresql: update failed: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("postgresql: failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return adapter.ErrNotFound
		}
	}

	return nil
}

// Delete removes records from the database.
func (a *PostgreSQLAdapter) Delete(ctx context.Context, op *adapter.Operation, identifiers []interface{}) error {
	if a.db == nil {
		return fmt.Errorf("postgresql: not connected")
	}

	query := op.Statement
	for _, id := range identifiers {
		var params map[string]interface{}
		if idMap, ok := id.(map[string]interface{}); ok {
			params = idMap
		} else {
			params = map[string]interface{}{"id": id}
		}

		args, err := extractArgs(query, params)
		if err != nil {
			return err
		}
		pgQuery := replaceNamedParams(query)

		result, err := a.db.ExecContext(ctx, pgQuery, args...)
		if err != nil {
			return fmt.Errorf("postgresql: delete failed: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("postgresql: failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return adapter.ErrNotFound
		}
	}

	return nil
}

// Execute runs custom SQL statements or stored procedures.
func (a *PostgreSQLAdapter) Execute(ctx context.Context, action *adapter.Action, params map[string]interface{}) (interface{}, error) {
	if a.db == nil {
		return nil, fmt.Errorf("postgresql: not connected")
	}

	query := action.Statement
	args, err := extractArgs(query, params)
	if err != nil {
		return nil, err
	}
	query = replaceNamedParams(query)

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgresql: execute failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("postgresql: failed to get columns: %w", err)
	}

	// Scan results
	var results []interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("postgresql: scan failed: %w", err)
		}

		// Build result map
		result := make(map[string]interface{})
		for i, col := range columns {
			result[col] = values[i]
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

// Helper functions

func getStringConfig(config map[string]interface{}, key, defaultVal string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultVal
}

func getIntConfig(config map[string]interface{}, key string, defaultVal int) int {
	if val, ok := config[key].(int); ok {
		return val
	}
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

// extractArgs extracts argument values from params based on named parameters in query
func extractArgs(query string, params map[string]interface{}) ([]interface{}, error) {
	args := []interface{}{}
	paramNames := []string{}

	// Find all {param} placeholders
	inBrace := false
	paramName := ""
	for _, ch := range query {
		if ch == '{' {
			inBrace = true
			paramName = ""
		} else if ch == '}' && inBrace {
			inBrace = false
			paramNames = append(paramNames, paramName)
		} else if inBrace {
			paramName += string(ch)
		}
	}

	// Extract values in order
	for _, name := range paramNames {
		val, ok := params[name]
		if !ok {
			return nil, fmt.Errorf("postgresql: missing parameter: %s", name)
		}
		args = append(args, val)
	}

	return args, nil
}

// replaceNamedParams converts {param} syntax to PostgreSQL $1, $2, ... syntax
func replaceNamedParams(query string) string {
	result := ""
	inBrace := false
	paramIndex := 1

	for _, ch := range query {
		if ch == '{' {
			inBrace = true
			result += fmt.Sprintf("$%d", paramIndex)
			paramIndex++
		} else if ch == '}' && inBrace {
			inBrace = false
		} else if !inBrace {
			result += string(ch)
		}
	}

	return result
}
