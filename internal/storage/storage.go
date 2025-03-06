package storage

import (
	"database/sql"
	"errors"
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
