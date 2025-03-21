package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type ShortenTextFile struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      int    `json:"user_id"`
}

func LoadLinksFromFile(ctx context.Context, store Storage, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Если файла нет, просто продолжаем работу
		}
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

		uuid := strconv.Itoa(store.Len(ctx))
		userID := link.UserID

		store.Save(ctx, uuid, link.ShortURL, link.OriginalURL, userID)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	return nil
}
