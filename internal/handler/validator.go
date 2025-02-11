package handler

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/devlongs/validator-service/internal/service"

	"github.com/go-chi/chi/v5"
)

type ValidatorHandler struct {
	service service.ValidatorService
}

func NewValidatorHandler(s service.ValidatorService) *ValidatorHandler {
	return &ValidatorHandler{service: s}
}

type createValidatorRequest struct {
	NumValidators int    `json:"num_validators"`
	FeeRecipient  string `json:"fee_recipient"`
}

type createValidatorResponse struct {
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
}

func (h *ValidatorHandler) CreateValidatorRequest(w http.ResponseWriter, r *http.Request) {
	var req createValidatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Basic validations.
	if req.NumValidators <= 0 {
		http.Error(w, "num_validators must be positive", http.StatusBadRequest)
		return
	}
	if !isValidEthereumAddress(req.FeeRecipient) {
		http.Error(w, "Invalid Ethereum address", http.StatusBadRequest)
		return
	}

	requestID, err := h.service.CreateValidatorRequest(req.NumValidators, req.FeeRecipient)
	if err != nil {
		http.Error(w, "Failed to create validator request", http.StatusInternalServerError)
		return
	}

	resp := createValidatorResponse{
		RequestID: requestID,
		Message:   "Validator creation in progress",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ValidatorHandler) GetValidatorStatus(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "requestID")
	if requestID == "" {
		http.Error(w, "requestID is required", http.StatusBadRequest)
		return
	}

	vr, keys, err := h.service.GetValidatorStatus(requestID)
	if err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if vr.Status == "successful" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": vr.Status,
			"keys":   keys,
		})
	} else if vr.Status == "failed" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  vr.Status,
			"message": "Error processing request",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  vr.Status,
			"message": "Validator creation is in progress",
		})
	}
}

// isValidEthereumAddress validates that the address starts with "0x" followed by 40 hex characters.
func isValidEthereumAddress(address string) bool {
	re := regexp.MustCompile("^0x[a-fA-F0-9]{40}$")
	return re.MatchString(address)
}
