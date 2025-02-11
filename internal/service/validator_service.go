package service

import (
	"errors"
	"math/rand"
	"time"

	"github.com/devlongs/validator-service/internal/repository"

	"github.com/google/uuid"
)

type ValidatorService interface {
	CreateValidatorRequest(numValidators int, feeRecipient string) (string, error)
	GetValidatorStatus(requestID string) (*repository.ValidatorRequest, []string, error)
	ProcessValidatorCreation(requestID string, numValidators int, feeRecipient string)
}

type validatorService struct {
	repo repository.ValidatorRepository
}

func NewValidatorService(repo repository.ValidatorRepository) ValidatorService {
	return &validatorService{repo: repo}
}

func (s *validatorService) CreateValidatorRequest(numValidators int, feeRecipient string) (string, error) {
	requestID := uuid.New().String()
	vr := &repository.ValidatorRequest{
		RequestID:     requestID,
		CreatedAt:     time.Now(),
		Status:        "started",
		NumValidators: numValidators,
		FeeRecipient:  feeRecipient,
	}
	if err := s.repo.CreateRequest(vr); err != nil {
		return "", err
	}
	// Spawn asynchronous task to simulate key generation.
	go s.ProcessValidatorCreation(requestID, numValidators, feeRecipient)
	return requestID, nil
}

func (s *validatorService) GetValidatorStatus(requestID string) (*repository.ValidatorRequest, []string, error) {
	return s.repo.GetRequest(requestID)
}

// keyGen simulates key generation with a 20ms delay per key.
func keyGen(numKeys int) ([]string, error) {
	keys := make([]string, 0, numKeys)
	for i := 0; i < numKeys; i++ {
		time.Sleep(20 * time.Millisecond)
		keys = append(keys, uuid.New().String())
	}
	// Simulate a 5% failure rate.
	if rand.Intn(100) < 5 {
		return nil, errors.New("simulated key generation error")
	}
	return keys, nil
}

func (s *validatorService) ProcessValidatorCreation(requestID string, numValidators int, feeRecipient string) {
	keys, err := keyGen(numValidators)
	if err != nil {
		s.repo.UpdateRequestStatus(requestID, "failed")
		return
	}
	if err := s.repo.AddValidatorKeys(requestID, keys, feeRecipient); err != nil {
		s.repo.UpdateRequestStatus(requestID, "failed")
		return
	}
	s.repo.UpdateRequestStatus(requestID, "successful")
}
