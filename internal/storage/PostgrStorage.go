package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

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
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_id INT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		correlation_id TEXT UNIQUE NOT NULL,
		short_url TEXT UNIQUE NOT NULL,
		original_url TEXT UNIQUE NOT NULL,
		user_id INT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
    `
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStorage) Save(ctx context.Context, correlationID string, short string, original string, userID int) (string, error) {
	var existingShortURL string

	err := s.db.QueryRow(
		`INSERT INTO urls (correlation_id, short_url, original_url, user_id) 
         VALUES ($1, $2, $3, $4) 
         ON CONFLICT (original_url) DO NOTHING 
         RETURNING short_url`,
		correlationID, short, original, userID,
	).Scan(&existingShortURL)

	// Если в `existingShortURL` пусто — значит, запись уже была, и нам нужно ее найти
	if errors.Is(err, sql.ErrNoRows) || existingShortURL == "" {
		existingShortURL, dbErr := s.GetFromOriginal(ctx, original)
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

func (s *PostgresStorage) Get(ctx context.Context, shortURL string) (string, bool, error) {
	var originalURL string
	err := s.db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url=$1", shortURL).Scan(&originalURL)
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

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *PostgresStorage) Len(ctx context.Context) int {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM urls").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (s *PostgresStorage) GetFromOriginal(ctx context.Context, originalURL string) (string, error) {
	var shorten string
	err := s.db.QueryRowContext(ctx, "SELECT short_url FROM urls WHERE original_url=$1", originalURL).Scan(&shorten)

	if err != nil {
		return "", err
	}
	link := shorten
	return link, nil
}

func (s *PostgresStorage) SaveUser(ctx context.Context, userID int) error {
	err := s.db.QueryRowContext(ctx, "INSERT INTO users (user_id) VALUES ($1)", userID)
	if err != nil {
		return err.Err()
	}
	return nil
}

func (s *PostgresStorage) GetUserFromID(ctx context.Context, userID int) (bool, error) {
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE user_id = $1", userID).Scan(&userID)
	if err != nil {
		return false, err
	}
	return true, err
}

func (s *PostgresStorage) GetNewUser(ctx context.Context) (int, error) {
	var countUsers int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(user_id) FROM users").Scan(&countUsers)
	if err != nil {
		return 0, err
	}
	return countUsers + 1, nil
}

func (s *PostgresStorage) GetLinksByUserID(ctx context.Context, userID int) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT short_url, original_url FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make(map[string]string)
	for rows.Next() {
		var shortURL, originalURL string
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return nil, err
		}
		links[shortURL] = originalURL
	}

	// Проверяем, была ли ошибка во время итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
