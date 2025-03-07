package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrURLAlreadyExists = errors.New("URL уже существует в базе данных")
	ErrUserNotFound     = errors.New("пользователь не найден")
)

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
	}

	LinkStorage struct {
		links     map[string]string
		users     map[int]bool
		userLinks map[int][]string
	}
	PostgresStorage struct {
		db *sql.DB
	}
)
