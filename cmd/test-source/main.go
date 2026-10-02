package main

import (
	"context"
	"fmt"
	"log"

	"myPersonalTechRadar/internal/database"
	"myPersonalTechRadar/internal/sources"
)

func main() {
	conn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	repository := sources.NewRepository(conn)

	source := sources.Source{
		Name:       "Go Blog",
		URL:        "https://go.dev/blog/",
		SourceType: "rss_feed",
	}

	created, err := repository.Create(context.Background(), source)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created source: %+v\n", created)
}
