package storage

import (
	"database/sql"
	"errors"
)

type (
	Storage interface {
		Save(correlationID string, short string, original string) error
		Get(original string) (string, bool, error)
		Len() int
		Ping() error
	}
	LinkStorage struct {
		links map[string]string
	}
	PostgresStorage struct {
		db *sql.DB
	}
)

func NewLinkStorage() *LinkStorage {
	return &LinkStorage{links: map[string]string{}}
}

func (s *LinkStorage) Save(correlationID, short string, original string) error {
	s.links[short] = original
	return nil
}

func (s *LinkStorage) Get(short string) (string, bool, error) {
	original, exists := s.links[short]
	return original, exists, nil
}

func (s *LinkStorage) Len() int {
	return len(s.links)
}

func (s *LinkStorage) Ping() error {
	return nil
}
func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	storage := &PostgresStorage{db: db}

	// Создаём таблицу, если её нет
	err = storage.createSchema()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *PostgresStorage) createSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		correlation_id TEXT UNIQUE NOT NULL,
		short_url TEXT UNIQUE NOT NULL,
		original_url TEXT NOT NULL
	);
    `
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStorage) Save(correlationID, shortURL, originalURL string) error {
	_, err := s.db.Exec("INSERT INTO urls (correlation_id, short_url, original_url) VALUES ($1, $2, $3)", correlationID, shortURL, originalURL)
	return err
}

func (s *PostgresStorage) Get(shortURL string) (string, bool, error) {
	var originalURL string
	err := s.db.QueryRow("SELECT original_url FROM urls WHERE short_url=$1", shortURL).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, errors.New("url not found")
		}
		return "", false, err
	}
	if originalURL == "" {
		return "", false, err
	}
	return originalURL, true, nil
}

func (s *PostgresStorage) Ping() error {
	return s.db.Ping()
}

func (s *PostgresStorage) Len() int {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
