package main

import (
	"flag"
	"os"
	"strings"
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&flagBaseURL, "b", "http://localhost:8080/", "base URL for shortened links")
	flag.StringVar(&flagPathToSave, "f", "default.txt", "Path to save urls JSON")
	flag.StringVar(&flagForDB, "d", "", "Postgresql info for connect")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		flagBaseURL = envBaseURL
	}
	if envPathToSave := os.Getenv("FILE_STORAGE_PATH"); envPathToSave != "" {
		flagPathToSave = envPathToSave
	}
	if envDBtoSave := os.Getenv("DATABASE_DSN"); envDBtoSave != "" {
		flagForDB = envDBtoSave
	}

	if !strings.HasSuffix(flagBaseURL, "/") {
		flagBaseURL += "/"
	}
}
