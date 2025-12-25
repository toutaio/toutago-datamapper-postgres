package main

import (
"context"
"fmt"
"log"

"github.com/toutago/toutago-datamapper/adapter"
"github.com/toutago/toutago-datamapper/config"
"github.com/toutago/toutago-datamapper/engine"
postgresql "github.com/toutago/toutago-datamapper-postgresql"
)

func main() {
fmt.Println("=== PostgreSQL Adapter Basic Example ===")

mapper, err := engine.NewMapper("config.yaml")
if err != nil {
log.Fatalf("Failed to create mapper: %v", err)
}
defer mapper.Close()

mapper.RegisterAdapter("postgresql", func(source config.Source) (adapter.Adapter, error) {
return postgresql.NewPostgreSQLAdapter(), nil
})

ctx := context.Background()

newUser := map[string]interface{}{
"Name":  "Alice Johnson",
"Email": "alice@example.com",
}

fmt.Printf("Creating user: %s\n", newUser["Name"])
if err := mapper.Insert(ctx, "users.insert", newUser); err != nil {
log.Fatalf("Insert failed: %v", err)
}
fmt.Printf("✓ User created with ID: %v\n", newUser["ID"])

var fetchedUser map[string]interface{}
if err := mapper.Fetch(ctx, "users.fetch", map[string]interface{}{
"id": newUser["ID"],
}, &fetchedUser); err != nil {
log.Fatalf("Fetch failed: %v", err)
}
fmt.Printf("✓ User fetched: %s\n", fetchedUser["Name"])

fmt.Println("\n=== Example Complete ===")
}
