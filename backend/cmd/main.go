package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"fixy-backend/internal/db"
	"fixy-backend/internal/router"
)

func main() {
	ctx := context.Background()
	database, err := db.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Printf("database unavailable: %v", err)
	}
	if database != nil {
		defer func() {
			if closeErr := database.Close(); closeErr != nil {
				log.Printf("close database: %v", closeErr)
			}
		}()
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router.New(database),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
