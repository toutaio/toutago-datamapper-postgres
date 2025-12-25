package postgresql

import (
	"context"
	"testing"

	"github.com/toutago/toutago-datamapper/adapter"
)

func TestPostgreSQLAdapter_Name(t *testing.T) {
	a := NewPostgreSQLAdapter()
	if name := a.Name(); name != "postgresql" {
		t.Errorf("expected name 'postgresql', got '%s'", name)
	}
}

func TestPostgreSQLAdapter_NewAdapter(t *testing.T) {
	a := NewPostgreSQLAdapter()
	if a == nil {
		t.Fatal("expected adapter instance, got nil")
	}
	if a.maxConn != 10 {
		t.Errorf("expected maxConn=10, got %d", a.maxConn)
	}
	if a.maxIdle != 5 {
		t.Errorf("expected maxIdle=5, got %d", a.maxIdle)
	}
	if a.connMaxAge != 3600 {
		t.Errorf("expected connMaxAge=3600, got %d", a.connMaxAge)
	}
}

func TestPostgreSQLAdapter_CloseWithoutConnect(t *testing.T) {
	a := NewPostgreSQLAdapter()
	if err := a.Close(); err != nil {
		t.Errorf("expected no error on close without connect, got %v", err)
	}
}

func TestReplaceNamedParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single parameter",
			input:    "SELECT * FROM users WHERE id = {id}",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "multiple parameters",
			input:    "SELECT * FROM users WHERE name = {name} AND email = {email}",
			expected: "SELECT * FROM users WHERE name = $1 AND email = $2",
		},
		{
			name:     "no parameters",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "parameter in different positions",
			input:    "INSERT INTO users (id, name) VALUES ({id}, {name})",
			expected: "INSERT INTO users (id, name) VALUES ($1, $2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceNamedParams(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractArgs(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		params    map[string]interface{}
		expected  []interface{}
		expectErr bool
	}{
		{
			name:      "single parameter",
			query:     "SELECT * FROM users WHERE id = {id}",
			params:    map[string]interface{}{"id": 123},
			expected:  []interface{}{123},
			expectErr: false,
		},
		{
			name:      "multiple parameters",
			query:     "SELECT * FROM users WHERE name = {name} AND email = {email}",
			params:    map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
			expected:  []interface{}{"Alice", "alice@example.com"},
			expectErr: false,
		},
		{
			name:      "missing parameter",
			query:     "SELECT * FROM users WHERE id = {id}",
			params:    map[string]interface{}{},
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "no parameters",
			query:     "SELECT * FROM users",
			params:    map[string]interface{}{},
			expected:  []interface{}{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractArgs(tt.query, tt.params)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d args, got %d", len(tt.expected), len(result))
				return
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Errorf("arg %d: expected %v, got %v", i, exp, result[i])
				}
			}
		})
	}
}

func TestPostgreSQLAdapter_FetchWithoutConnect(t *testing.T) {
	a := NewPostgreSQLAdapter()
	ctx := context.Background()
	op := &adapter.Operation{
		Statement: "SELECT id FROM users WHERE id = {id}",
	}
	params := map[string]interface{}{"id": 1}

	_, err := a.Fetch(ctx, op, params)
	if err == nil {
		t.Error("expected error when not connected, got nil")
	}
}

func TestPostgreSQLAdapter_InsertWithoutConnect(t *testing.T) {
	a := NewPostgreSQLAdapter()
	ctx := context.Background()
	op := &adapter.Operation{
		Statement: "users",
	}
	objects := []interface{}{
		map[string]interface{}{"name": "test"},
	}

	err := a.Insert(ctx, op, objects)
	if err == nil {
		t.Error("expected error when not connected, got nil")
	}
}

func TestPostgreSQLAdapter_UpdateWithoutConnect(t *testing.T) {
	a := NewPostgreSQLAdapter()
	ctx := context.Background()
	op := &adapter.Operation{
		Statement: "UPDATE users SET name = {name} WHERE id = {id}",
	}
	objects := []interface{}{
		map[string]interface{}{"id": 1, "name": "test"},
	}

	err := a.Update(ctx, op, objects)
	if err == nil {
		t.Error("expected error when not connected, got nil")
	}
}

func TestPostgreSQLAdapter_DeleteWithoutConnect(t *testing.T) {
	a := NewPostgreSQLAdapter()
	ctx := context.Background()
	op := &adapter.Operation{
		Statement: "DELETE FROM users WHERE id = {id}",
	}

	err := a.Delete(ctx, op, []interface{}{1})
	if err == nil {
		t.Error("expected error when not connected, got nil")
	}
}

func TestPostgreSQLAdapter_ExecuteWithoutConnect(t *testing.T) {
	a := NewPostgreSQLAdapter()
	ctx := context.Background()
	action := &adapter.Action{
		Statement: "SELECT COUNT(*) as count FROM users",
	}

	_, err := a.Execute(ctx, action, nil)
	if err == nil {
		t.Error("expected error when not connected, got nil")
	}
}

func TestGetStringConfig(t *testing.T) {
	tests := []struct {
		name       string
		config     map[string]interface{}
		key        string
		defaultVal string
		expected   string
	}{
		{
			name:       "key exists",
			config:     map[string]interface{}{"host": "localhost"},
			key:        "host",
			defaultVal: "default",
			expected:   "localhost",
		},
		{
			name:       "key missing",
			config:     map[string]interface{}{},
			key:        "host",
			defaultVal: "default",
			expected:   "default",
		},
		{
			name:       "wrong type",
			config:     map[string]interface{}{"host": 123},
			key:        "host",
			defaultVal: "default",
			expected:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringConfig(tt.config, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetIntConfig(t *testing.T) {
	tests := []struct {
		name       string
		config     map[string]interface{}
		key        string
		defaultVal int
		expected   int
	}{
		{
			name:       "key exists as int",
			config:     map[string]interface{}{"port": 5432},
			key:        "port",
			defaultVal: 3306,
			expected:   5432,
		},
		{
			name:       "key exists as float64",
			config:     map[string]interface{}{"port": 5432.0},
			key:        "port",
			defaultVal: 3306,
			expected:   5432,
		},
		{
			name:       "key missing",
			config:     map[string]interface{}{},
			key:        "port",
			defaultVal: 3306,
			expected:   3306,
		},
		{
			name:       "wrong type",
			config:     map[string]interface{}{"port": "5432"},
			key:        "port",
			defaultVal: 3306,
			expected:   3306,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIntConfig(tt.config, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
