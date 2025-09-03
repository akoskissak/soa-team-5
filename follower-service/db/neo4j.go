package db

import (
	"context"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var Driver neo4j.DriverWithContext

func ConnectNeo4j(uri, user, pass string) {
	if uri == "" || user == "" || pass == "" {
		log.Fatal("NEO4J env variables not set")
	}

	var drv neo4j.DriverWithContext
	var err error
	maxRetries := 5
	retryDelay := 3 * time.Second

	for i := 1; i <= maxRetries; i++ {
		drv, err = neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, pass, ""))
		if err != nil {
			log.Printf("[WARN] Attempt %d: Failed to create Neo4j driver: %v\n", i, err)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := drv.VerifyConnectivity(ctx); err != nil {
				log.Printf("[WARN] Attempt %d: Neo4j not reachable: %v\n", i, err)
			} else {
				Driver = drv
				log.Println("[INFO] Neo4j connected!")
				return
			}
		}

		if i < maxRetries {
			log.Printf("[INFO] Retrying in %s...\n", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	log.Fatal("[ERROR] Could not connect to Neo4j after multiple attempts:", err)
}
