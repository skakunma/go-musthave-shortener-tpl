package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrURLAlreadyExists = errors.New("URL уже существует в базе данных")
	ErrUserNotFound     = errors.New("Пользователь не найден")
)

type (
	Storage interface {
		Save(ctx context.Context, correlationID string, short string, original string, userId int) (string, error)
		Get(ctx context.Context, original string) (string, bool, error)
		Len(ctx context.Context) int
		Ping(ctx context.Context) error
		GetFromOriginal(ctx context.Context, original string) (string, error)
		SaveUser(ctx context.Context, userId int) error
		GetUserFromID(ctx context.Context, userId int) (bool, error)
		GetNewUser(ctx context.Context) (int, error)
		GetLinksByUserID(ctx context.Context, userId int) (map[string]string, error)
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
