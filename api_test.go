package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devlongs/validator-service/internal/db"
	"github.com/devlongs/validator-service/internal/handler"
	"github.com/devlongs/validator-service/internal/repository"
	"github.com/devlongs/validator-service/internal/service"
	"github.com/go-chi/chi/v5"
)

// setupRouter creates an in-memory database and returns a configured router.
func setupRouter(t *testing.T) *chi.Mux {
	// Use an in-memory SQLite DB for testing.
	dbConn, err := db.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("Error creating in-memory DB: %v", err)
	}
	repo := repository.NewValidatorRepository(dbConn)
	svc := service.NewValidatorService(repo)
	validatorHandler := handler.NewValidatorHandler(svc)
	healthHandler := handler.NewHealthHandler(dbConn)

	r := chi.NewRouter()
	r.Get("/health", healthHandler.HealthCheck)
	r.Post("/validators", validatorHandler.CreateValidatorRequest)
	r.Get("/validators/{requestID}", validatorHandler.GetValidatorStatus)
	return r
}

// TestHealthCheck verifies that the /health endpoint returns a healthy status.
func TestHealthCheck(t *testing.T) {
	router := setupRouter(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to send GET /health: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if status, ok := result["status"]; !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got: %v", result)
	}
}

// TestValidatorFlow tests the creation of a validator request and then retrieves its status.
func TestValidatorFlow(t *testing.T) {
	// Set a deterministic seed to reduce randomness in tests.
	rand.Seed(42)

	router := setupRouter(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Define the payload for creating a validator request.
	payload := map[string]interface{}{
		"num_validators": 3,
		"fee_recipient":  "0x1234567890abcdef1234567890abcdef12345678",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Send POST request to /validators.
	resp, err := http.Post(ts.URL+"/validators", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to send POST /validators: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read POST response: %v", err)
	}

	var createResp map[string]string
	if err := json.Unmarshal(body, &createResp); err != nil {
		t.Fatalf("Failed to unmarshal create response: %v", err)
	}

	requestID, ok := createResp["request_id"]
	if !ok || requestID == "" {
		t.Fatalf("Invalid request_id in response: %v", createResp)
	}

	// Wait for the asynchronous key generation to (hopefully) complete.
	time.Sleep(200 * time.Millisecond)

	// Send GET request to /validators/{requestID}.
	getResp, err := http.Get(ts.URL + "/validators/" + requestID)
	if err != nil {
		t.Fatalf("Failed to send GET /validators/{requestID}: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200 for GET, got: %d", getResp.StatusCode)
	}

	getBody, err := ioutil.ReadAll(getResp.Body)
	if err != nil {
		t.Fatalf("Failed to read GET response: %v", err)
	}

	var statusResp map[string]interface{}
	if err := json.Unmarshal(getBody, &statusResp); err != nil {
		t.Fatalf("Failed to unmarshal GET response: %v", err)
	}

	status, ok := statusResp["status"].(string)
	if !ok {
		t.Fatalf("Response does not contain a valid status: %v", statusResp)
	}

	// Verify response based on the status value.
	if status == "successful" {
		keys, ok := statusResp["keys"].([]interface{})
		if !ok {
			t.Errorf("Expected 'keys' to be an array, got: %v", statusResp["keys"])
		}
		if len(keys) != 3 {
			t.Errorf("Expected 3 keys, got: %d", len(keys))
		}
	} else if status == "failed" {
		msg, _ := statusResp["message"].(string)
		if msg != "Error processing request" {
			t.Errorf("Expected failure message 'Error processing request', got: %s", msg)
		}
	} else {
		t.Errorf("Unexpected status: %s", status)
	}
}
