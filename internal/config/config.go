package config

import (
	"GoIncrease1/internal/storage"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

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

func LoadLinksFromFile(ctx context.Context) error {
	file, err := os.Open(Cfg.FlagPathToSave)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var link ShortenTextFile
		err := json.Unmarshal(scanner.Bytes(), &link)
		if err != nil {
			return fmt.Errorf("failed to parse JSON: %v", err)
		}
		uuid := strconv.Itoa(Cfg.Store.Len(ctx) - 1)
		userID := link.UserID

		Cfg.Store.Save(ctx, uuid, link.ShortURL, link.OriginalURL, userID)

	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	return nil
}

func NewConfig() *Config {
	return &Config{
		Charset:       "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		CharsetLength: 7,
	}
}
