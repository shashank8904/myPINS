package main

import (
	"context"
	"fmt"
	"log"

	"myPersonalTechRadar/internal/database"
)

func main() {
	ctx := context.Background()
	conn, err := database.Connect()
	if err != nil {
		log.Fatalf("database connect: %v", err)
	}
	defer conn.Close()

	// Seed some topics
	queries := []string{
		`INSERT INTO topics (name, slug) VALUES ('Kubernetes', 'kubernetes') ON CONFLICT DO NOTHING;`,
		`INSERT INTO topics (name, slug) VALUES ('Go', 'go') ON CONFLICT DO NOTHING;`,
		`INSERT INTO topics (name, slug) VALUES ('Rust', 'rust') ON CONFLICT DO NOTHING;`,

		// Seed some interests (Assume Go is 0.9, Kubernetes is 0.5)
		`INSERT INTO interests (topic_id, weight) 
		 SELECT id, 0.9 FROM topics WHERE name = 'Go'
		 ON CONFLICT DO NOTHING;`,

		`INSERT INTO interests (topic_id, weight) 
		 SELECT id, 0.5 FROM topics WHERE name = 'Kubernetes'
		 ON CONFLICT DO NOTHING;`,

		// Seed roadmap (Assume Kubernetes is 'current', Rust is 'next')
		`INSERT INTO roadmap_items (topic_id, status) 
		 SELECT id, 'current' FROM topics WHERE name = 'Kubernetes'
		 ON CONFLICT DO NOTHING;`,

		`INSERT INTO roadmap_items (topic_id, status) 
		 SELECT id, 'next' FROM topics WHERE name = 'Rust'
		 ON CONFLICT DO NOTHING;`,
	}

	for _, q := range queries {
		_, err := conn.Exec(ctx, q)
		if err != nil {
			log.Printf("Query failed: %v\nQuery: %s", err, q)
		}
	}

	fmt.Println("Seeded personalization data.")
}
