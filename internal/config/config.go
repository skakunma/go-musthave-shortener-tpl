package config

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/storage"

	"go.uber.org/zap"
)

var Cfg *Config

type (
	Config struct {
		Mu             sync.Mutex
		Charset        string
		CharsetLength  int
		Sugar          *zap.SugaredLogger
		File           *os.File
		FlagRunAddr    string
		FlagBaseURL    string
		FlagPathToSave string
		FlagForDB      string
		Store          storage.Storage
	}
)

func LoadConfig(ctx context.Context) (*Config, error) {
	cfg := &Config{
		Charset:       "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		CharsetLength: 7,
	}

	// Инициализация логгера
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации логгера: %w", err)
	}
	cfg.Sugar = logger.Sugar()
	ParseFlags(cfg)
	if cfg.FlagForDB != "" {
		pgStorage, err := storage.NewPostgresStorage(cfg.FlagForDB)
		if err != nil {
			cfg.Sugar.Error("Ошибка подключения к БД:", err)
		} else {
			cfg.Store = pgStorage
		}
	} else {
		cfg.Store = storage.NewLinkStorage()
	}

	cfg.File, err = os.OpenFile(cfg.FlagPathToSave, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		cfg.Sugar.Errorf("Ошибка открытия файла: %v", err)
	}

	if err := storage.LoadLinksFromFile(ctx, cfg.Store, cfg.FlagPathToSave); err != nil {
		cfg.Sugar.Error("Ошибка загрузки ссылок:", err)
	}

	return cfg, nil
}
