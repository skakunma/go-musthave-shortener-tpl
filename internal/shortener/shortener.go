package shortener

import (
	"GoIncrease1/internal/config"
	"GoIncrease1/internal/storage"

	"encoding/json"
	"errors"
	"math/rand"
	"strings"
)

var ErrURLAlreadyExists = errors.New("URL уже существует в базе данных")

type (
	ShortenTextFile struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)

func (info *ShortenTextFile) SaveURLInfo() error {
	encoder := json.NewEncoder(config.Cfg.File)
	err := encoder.Encode(info)
	if err != nil {
		return err
	}
	return nil
}

func generateLink() string {
	var builder strings.Builder
	builder.Grow(config.Cfg.CharsetLength)

	for i := 0; i < config.Cfg.CharsetLength; i++ {
		indx := rand.Intn(len(config.Cfg.Charset) - 1)
		builder.WriteByte(config.Cfg.Charset[indx])
	}

	return builder.String()
}
func AddLink(Link string, uuid string) (string, error) {
	config.Cfg.Mu.Lock()
	defer config.Cfg.Mu.Unlock()

	for {
		randomLink := generateLink()

		if _, exist, _ := config.Cfg.Store.Get(randomLink); !exist {
			shortenLink, err := config.Cfg.Store.Save(uuid, randomLink, Link)
			if err != nil {
				if errors.Is(err, storage.ErrURLAlreadyExists) {
					return config.Cfg.FlagBaseURL + shortenLink, err
				}
				return "", err
			}

			url := ShortenTextFile{UUID: uuid, ShortURL: randomLink, OriginalURL: Link}
			err = url.SaveURLInfo()
			if err != nil {
				return "", err
			}
			return config.Cfg.FlagBaseURL + shortenLink, nil
		}
	}
}

func GetLink(key string) (string, bool) {
	if value, exist, err := config.Cfg.Store.Get(key); exist && err == nil {
		return value, true
	}

	return "", false
}
