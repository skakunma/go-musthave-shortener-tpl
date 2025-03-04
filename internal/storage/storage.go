package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrURLAlreadyExists = errors.New("URL уже существует в базе данных")

type (
	Storage interface {
		Save(correlationID string, short string, original string) (string, error)
		Get(original string) (string, bool, error)
		Len() int
		Ping() error
		GetFromOriginal(string) (string, error)
	}

	LinkStorage struct {
		links map[string]string
	}
	PostgresStorage struct {
		db *sql.DB
	}
)

func (s *LinkStorage) GetFromOriginal(originalURL string) (string, error) {
	return originalURL, nil
}

func NewLinkStorage() *LinkStorage {

	return &LinkStorage{links: map[string]string{}}
}

func (s *LinkStorage) Save(correlationID string, short string, original string) (string, error) {
	s.links[short] = original
	return short, nil
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
		original_url TEXT UNIQUE NOT NULL
	);
    `
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStorage) Save(correlationID string, short string, original string) (string, error) {
	var existingShortURL string

	err := s.db.QueryRow(
		`INSERT INTO urls (correlation_id, short_url, original_url) 
         VALUES ($1, $2, $3) 
         ON CONFLICT (original_url) DO NOTHING 
         RETURNING short_url`,
		correlationID, short, original,
	).Scan(&existingShortURL)

	// Если в `existingShortURL` пусто — значит, запись уже была, и нам нужно ее найти
	if errors.Is(err, sql.ErrNoRows) || existingShortURL == "" {
		existingShortURL, dbErr := s.GetFromOriginal(original)
		if dbErr != nil {
			return "", fmt.Errorf("ошибка получения существующего URL: %w", dbErr)
		}
		return existingShortURL, ErrURLAlreadyExists
	}

	if err != nil {
		return "", fmt.Errorf("ошибка сохранения в БД: %w", err)
	}

	return existingShortURL, nil
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

func (s *PostgresStorage) GetFromOriginal(originalURL string) (string, error) {
	var shorten string
	err := s.db.QueryRow("SELECT short_url FROM urls WHERE original_url=$1", originalURL).Scan(&shorten)
	if err != nil {
		return "", err
	}
	link := shorten
	return link, nil
}
