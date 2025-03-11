package config

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/storage"
	"go.uber.org/zap"
)

var Cfg *Config

type (
	ShortenTextFile struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		UserID      int    `json:"user_id"`
	}
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

func LoadLinksFromFile(ctx context.Context, cfg *Config) error {

	file, err := os.Open(cfg.FlagPathToSave)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var link ShortenTextFile
		err := json.Unmarshal(scanner.Bytes(), &link)
		if err != nil {
			return fmt.Errorf("ошибка парсинга JSON: %w", err)
		}
		uuid := strconv.Itoa(cfg.Store.Len(ctx) - 1)
		userID := link.UserID

		cfg.Store.Save(ctx, uuid, link.ShortURL, link.OriginalURL, userID)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	return nil
}

func LoadConfig(ctx context.Context) (*Config, error) {
	cfg := &Config{
		Charset:       "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		CharsetLength: 7,
	}
	ParseFlags(cfg)

	// Инициализация логгера
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации логгера: %w", err)
	}
	cfg.Sugar = logger.Sugar()

	// Выбираем хранилище (PostgreSQL или in-memory)
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

	// Загружаем ссылки из файла
	if err := LoadLinksFromFile(ctx, cfg); err != nil {
		cfg.Sugar.Error("Ошибка загрузки ссылок:", err)
	}

	// Открываем файл для записи
	cfg.File, err = os.OpenFile(cfg.FlagPathToSave, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		cfg.Sugar.Errorf("Ошибка открытия файла: %v", err)
	}

	return cfg, nil
}
