package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// NewSQLiteDB opens or creates the SQLite database and sets up the schema.
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS validator_requests (
		request_id TEXT PRIMARY KEY,
		created_at DATETIME,
		status TEXT,
		num_validators INTEGER,
		fee_recipient TEXT
	);
	CREATE TABLE IF NOT EXISTS validator_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request_id TEXT,
		key TEXT,
		fee_recipient TEXT,
		FOREIGN KEY(request_id) REFERENCES validator_requests(request_id)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return db, nil
}
