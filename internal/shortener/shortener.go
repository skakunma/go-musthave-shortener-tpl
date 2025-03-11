package shortener

import (
	"context"

	"encoding/json"
	"errors"
	"math/rand"
	"strings"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"
)

var ErrURLAlreadyExists = errors.New("URL уже существует в базе данных")

type (
	ShortenTextFile struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		UserID      int    `json:"user_id"`
	}
)

func (info *ShortenTextFile) SaveURLInfo(cfg *config.Config) error {
	encoder := json.NewEncoder(cfg.File)
	err := encoder.Encode(info)
	if err != nil {
		return err
	}
	return nil
}

func GenerateLink(cfg *config.Config) string {
	var builder strings.Builder
	builder.Grow(cfg.CharsetLength)

	for i := 0; i < cfg.CharsetLength; i++ {
		indx := rand.Intn(len(cfg.Charset) - 1)
		builder.WriteByte(cfg.Charset[indx])
	}

	return builder.String()
}
func AddLink(ctx context.Context, cfg *config.Config, Link string, uuid string, UserID int) (string, error) {
	cfg.Mu.Lock()
	defer cfg.Mu.Unlock()

	for {
		randomLink := GenerateLink(cfg)

		if _, exist, _ := cfg.Store.Get(ctx, randomLink); !exist {
			shortenLink, err := cfg.Store.Save(ctx, uuid, randomLink, Link, UserID)
			if err != nil {
				if errors.Is(err, ErrURLAlreadyExists) {
					return cfg.FlagBaseURL + shortenLink, err
				}
				return "", err
			}

			url := ShortenTextFile{UUID: uuid, ShortURL: randomLink, OriginalURL: Link, UserID: UserID}
			err = url.SaveURLInfo(cfg)
			if err != nil {
				return "", err
			}
			return cfg.FlagBaseURL + shortenLink, nil
		}
	}
}

func GetLink(ctx context.Context, cfg *config.Config, key string) (string, bool) {
	if value, exist, err := cfg.Store.Get(ctx, key); exist && err == nil {
		return value, true
	}

	return "", false
}
