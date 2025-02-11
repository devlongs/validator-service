package main

import (
	"log"
	"net/http"
	"time"

	"github.com/devlongs/validator-service/internal/db"
	"github.com/devlongs/validator-service/internal/handler"
	"github.com/devlongs/validator-service/internal/repository"
	"github.com/devlongs/validator-service/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	sqliteDB, err := db.NewSQLiteDB("validators.db")
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}
	defer sqliteDB.Close()

	// Initialize repository (with dependency injection)
	repo := repository.NewValidatorRepository(sqliteDB)

	// Initialize service with the repository
	validatorService := service.NewValidatorService(repo)

	// Create handlers using the service and DB for health check
	validatorHandler := handler.NewValidatorHandler(validatorService)
	healthHandler := handler.NewHealthHandler(sqliteDB)

	// Setup router using chi
	r := chi.NewRouter()
	r.Get("/health", healthHandler.HealthCheck)
	r.Post("/validators", validatorHandler.CreateValidatorRequest)
	r.Get("/validators/{requestID}", validatorHandler.GetValidatorStatus)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("Server starting on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
