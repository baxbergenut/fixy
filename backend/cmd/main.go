package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"fixy-backend/internal/config"
	"fixy-backend/internal/db"
	"fixy-backend/internal/middleware"
	"fixy-backend/internal/router"
)

func main() {
	if err := config.Load(".env", "backend/.env", "../.env"); err != nil {
		log.Printf("load env: %v", err)
	}

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
		Handler:           middleware.CORS(middleware.Logger(router.New(database, os.Getenv("GROQ_TOKEN")))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
