package config

import (
	"GoIncrease1/internal/storage"
	"bufio"
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

func LoadLinksFromFile() error {
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
		uuid := strconv.Itoa(Cfg.Store.Len() - 1)

		Cfg.Store.Save(uuid, link.ShortURL, link.OriginalURL)

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
