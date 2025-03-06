package storage

import (
	"context"
	"database/sql"
	"errors"
)

var ErrURLAlreadyExists = errors.New("URL уже существует в базе данных")

type (
	Storage interface {
		Save(ctx context.Context, correlationID string, short string, original string) (string, error)
		Get(ctx context.Context, original string) (string, bool, error)
		Len(ctx context.Context) int
		Ping(ctx context.Context) error
		GetFromOriginal(ctx context.Context, original string) (string, error)
	}

	LinkStorage struct {
		links map[string]string
	}
	PostgresStorage struct {
		db *sql.DB
	}
)
