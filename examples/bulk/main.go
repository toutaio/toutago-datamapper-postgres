package main

import (
	"context"
	"fmt"
	"log"
	"time"

	postgresql "github.com/toutaio/toutago-datamapper-postgres"
	"github.com/toutaio/toutago-datamapper/adapter"
	"github.com/toutaio/toutago-datamapper/config"
	"github.com/toutaio/toutago-datamapper/engine"
)

func main() {
	fmt.Println("=== PostgreSQL Bulk Operations Example ===")

	mapper, err := engine.NewMapper("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create mapper: %v", err)
	}
	defer mapper.Close()

	mapper.RegisterAdapter("postgresql", func(source config.Source) (adapter.Adapter, error) {
		return postgresql.NewPostgreSQLAdapter(), nil
	})

	ctx := context.Background()

	products := []interface{}{
		map[string]interface{}{"Name": "Laptop", "Price": 1299.99, "Stock": 50},
		map[string]interface{}{"Name": "Mouse", "Price": 29.99, "Stock": 200},
		map[string]interface{}{"Name": "Keyboard", "Price": 149.99, "Stock": 75},
	}

	fmt.Printf("Bulk inserting %d products...\n", len(products))
	start := time.Now()
	if err := mapper.Insert(ctx, "products.bulk-insert", products); err != nil {
		log.Fatalf("Bulk insert failed: %v", err)
	}
	fmt.Printf("✓ Inserted in %v\n", time.Since(start))

	fmt.Println("\n=== Example Complete ===")
}
