package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"renet-catalog/config"
	"renet-catalog/db/repositories"
	"renet-catalog/services"
)

func main() {
	dbURL := config.GetString("DB_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "ERROR: DB_URL environment variable is required but not set.")
		os.Exit(1)
	}

	omdbAPIKey := config.GetString("OMDB_API_KEY")
	if omdbAPIKey == "" {
		fmt.Fprintln(os.Stderr, "ERROR: OMDB_API_KEY environment variable is required but not set.")
		os.Exit(1)
	}

	redisAddr := config.GetString("REDIS_ADDR")

	db, err := config.ConnectDB(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[ingest] Connected to PostgreSQL.")

	rdb := config.ConnectRedis(redisAddr)
	if rdb != nil {
		defer rdb.Close()
		fmt.Println("[ingest] Connected to Redis.")
	} else {
		fmt.Println("[ingest] Redis not available; skipped.")
	}

	omdbHTTPClient := &http.Client{Timeout: 10 * time.Second}

	repo := repositories.NewCatalogRepository(db)
	svc := services.NewCatalogServiceImpl(repo, omdbHTTPClient, omdbAPIKey, rdb)

	currentYear := time.Now().Year()
	years := []int{currentYear, currentYear - 1}
	fmt.Printf("[ingest] Discovering Indian cinema for years: %v\n", years)

	count, err := svc.DiscoverAndIngestIndianMovies(years)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Ingestion failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[ingest] Done. %d new Indian movies added.\n", count)
}