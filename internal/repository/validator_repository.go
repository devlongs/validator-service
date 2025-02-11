package repository

import (
	"database/sql"
	"time"
)

type ValidatorRequest struct {
	RequestID     string
	CreatedAt     time.Time
	Status        string
	NumValidators int
	FeeRecipient  string
}

type ValidatorRepository interface {
	CreateRequest(vr *ValidatorRequest) error
	UpdateRequestStatus(requestID string, status string) error
	AddValidatorKeys(requestID string, keys []string, feeRecipient string) error
	GetRequest(requestID string) (*ValidatorRequest, []string, error)
}

type validatorRepository struct {
	db *sql.DB
}

func NewValidatorRepository(db *sql.DB) ValidatorRepository {
	return &validatorRepository{db: db}
}

func (r *validatorRepository) CreateRequest(vr *ValidatorRequest) error {
	query := `INSERT INTO validator_requests (request_id, created_at, status, num_validators, fee_recipient) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, vr.RequestID, vr.CreatedAt, vr.Status, vr.NumValidators, vr.FeeRecipient)
	return err
}

func (r *validatorRepository) UpdateRequestStatus(requestID string, status string) error {
	query := `UPDATE validator_requests SET status = ? WHERE request_id = ?`
	_, err := r.db.Exec(query, status, requestID)
	return err
}

func (r *validatorRepository) AddValidatorKeys(requestID string, keys []string, feeRecipient string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO validator_keys (request_id, key, fee_recipient) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, key := range keys {
		if _, err := stmt.Exec(requestID, key, feeRecipient); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *validatorRepository) GetRequest(requestID string) (*ValidatorRequest, []string, error) {
	var vr ValidatorRequest
	query := `SELECT request_id, created_at, status, num_validators, fee_recipient FROM validator_requests WHERE request_id = ?`
	row := r.db.QueryRow(query, requestID)
	err := row.Scan(&vr.RequestID, &vr.CreatedAt, &vr.Status, &vr.NumValidators, &vr.FeeRecipient)
	if err != nil {
		return nil, nil, err
	}

	var keys []string
	if vr.Status == "successful" {
		keyQuery := `SELECT key FROM validator_keys WHERE request_id = ?`
		rows, err := r.db.Query(keyQuery, requestID)
		if err != nil {
			return &vr, nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var key string
			if err := rows.Scan(&key); err != nil {
				return &vr, nil, err
			}
			keys = append(keys, key)
		}
	}
	return &vr, keys, nil
}
