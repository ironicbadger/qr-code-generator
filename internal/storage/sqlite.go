package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type QRCode struct {
	ID        int64
	Content   string
	Label     string
	ImageData []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS qr_codes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		label TEXT DEFAULT '',
		image_data BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_created_at ON qr_codes(created_at DESC);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) Create(content string, label string, imageData []byte) (*QRCode, error) {
	result, err := s.db.Exec(
		"INSERT INTO qr_codes (content, label, image_data) VALUES (?, ?, ?)",
		content, label, imageData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert qr code: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.GetByID(id)
}

func (s *Store) GetByID(id int64) (*QRCode, error) {
	qr := &QRCode{}
	err := s.db.QueryRow(
		"SELECT id, content, label, image_data, created_at, updated_at FROM qr_codes WHERE id = ?",
		id,
	).Scan(&qr.ID, &qr.Content, &qr.Label, &qr.ImageData, &qr.CreatedAt, &qr.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get qr code: %w", err)
	}
	return qr, nil
}

func (s *Store) List(limit, offset int) ([]*QRCode, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(
		"SELECT id, content, label, image_data, created_at, updated_at FROM qr_codes ORDER BY created_at DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list qr codes: %w", err)
	}
	defer rows.Close()

	var codes []*QRCode
	for rows.Next() {
		qr := &QRCode{}
		if err := rows.Scan(&qr.ID, &qr.Content, &qr.Label, &qr.ImageData, &qr.CreatedAt, &qr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan qr code: %w", err)
		}
		codes = append(codes, qr)
	}
	return codes, rows.Err()
}

func (s *Store) UpdateLabel(id int64, label string) error {
	result, err := s.db.Exec(
		"UPDATE qr_codes SET label = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		label, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update label: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("qr code not found")
	}
	return nil
}

func (s *Store) Delete(id int64) error {
	result, err := s.db.Exec("DELETE FROM qr_codes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete qr code: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("qr code not found")
	}
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
