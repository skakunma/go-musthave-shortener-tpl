package main

import (
	"flag"
	"os"
	"strings"
)

var (
	flagRunAddr string
	flagBaseURL string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&flagBaseURL, "b", "http://localhost:8080/", "base URL for shortened links")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		flagBaseURL = envBaseURL
	}

	if !strings.HasSuffix(flagBaseURL, "/") {
		flagBaseURL += "/"
	}
}
