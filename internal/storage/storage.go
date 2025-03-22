package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrURLAlreadyExists = errors.New("URL уже существует в базе данных")
	ErrUserNotFound     = errors.New("пользователь не найден")
	ErrLinkIsDeleted    = errors.New("Link is deleted")
)

type InfoAboutURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	ShortLink     string
}

type (
	Storage interface {
		Save(ctx context.Context, correlationID string, short string, original string, userID int) (string, error)
		Get(ctx context.Context, original string) (string, bool, error)
		Len(ctx context.Context) int
		Ping(ctx context.Context) error
		GetFromOriginal(ctx context.Context, original string) (string, error)
		SaveUser(ctx context.Context, userID int) error
		GetUserFromID(ctx context.Context, userID int) (bool, error)
		GetNewUser(ctx context.Context) (int, error)
		GetLinksByUserID(ctx context.Context, userID int) (map[string]string, error)
		AddLinksBatch(ctx context.Context, links []InfoAboutURL, userID int) ([]string, error)
		DeleteURL(ctx context.Context, UUID string) error
		GetUserFromUUID(ctx context.Context, UUID string) (int, error)
	}

	LinkStorage struct {
		links        map[string]string
		users        map[int]bool
		userLinks    map[int][]string
		deletedLinks map[string]bool
		uuidUser     map[string]int
	}
	PostgresStorage struct {
		db *sql.DB
	}
)
