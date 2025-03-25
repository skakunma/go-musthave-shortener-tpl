package config

import (
	"flag"
	"os"
	"strings"
)

func ParseFlags(cfg *Config) {
	// Определяем флаги
	flag.StringVar(&cfg.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.FlagBaseURL, "b", "http://localhost:8080", "base URL for shortened links")
	flag.StringVar(&cfg.FlagPathToSave, "f", "default.txt", "Path to save urls JSON")
	flag.StringVar(&cfg.FlagForDB, "d", "", "PostgreSQL connection string")

	flag.Parse()

	// Перезаписываем значениями из переменных окружения (если они есть)
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		cfg.FlagRunAddr = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.FlagBaseURL = envBaseURL
	}
	if envPathToSave := os.Getenv("FILE_STORAGE_PATH"); envPathToSave != "" {
		cfg.FlagPathToSave = envPathToSave
	}
	if envDBtoSave := os.Getenv("DATABASE_DSN"); envDBtoSave != "" {
		cfg.FlagForDB = envDBtoSave
	}

	if cfg.FlagPathToSave == "" {
		cfg.FlagPathToSave = "default.txt"
	}

	if !strings.HasSuffix(cfg.FlagBaseURL, "/") {
		cfg.FlagBaseURL += "/"
	}
}
