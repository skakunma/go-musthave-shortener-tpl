package main

import (
	"GoIncrease1/internal/config"
	"flag"
	"os"
	"strings"
)

func parseFlags() {
	flag.StringVar(&config.Cfg.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.Cfg.FlagBaseURL, "b", "http://localhost:8080/", "base URL for shortened links")
	flag.StringVar(&config.Cfg.FlagPathToSave, "f", "default.txt", "Path to save urls JSON")
	flag.StringVar(&config.Cfg.FlagForDB, "d", "host=localhost user=postgres password=example dbname=postgres sslmode=disable", "Postgresql info for connect")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		config.Cfg.FlagRunAddr = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.Cfg.FlagBaseURL = envBaseURL
	}
	if envPathToSave := os.Getenv("FILE_STORAGE_PATH"); envPathToSave != "" {
		config.Cfg.FlagPathToSave = envPathToSave
	}
	if envDBtoSave := os.Getenv("DATABASE_DSN"); envDBtoSave != "" {
		config.Cfg.FlagForDB = envDBtoSave
	}

	if !strings.HasSuffix(config.Cfg.FlagBaseURL, "/") {
		config.Cfg.FlagBaseURL += "/"
	}
}
