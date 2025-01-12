package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var (
	mu      sync.Mutex
	Links   = make(map[string]string)
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func generateLink() string {
	var builder strings.Builder
	builder.Grow(7)

	for i := 0; i < 7; i++ {
		indx := rand.Intn(51)
		builder.WriteByte(charset[indx])
	}

	return builder.String()
}
func AddLink(Link string) string {
	for {
		randomLink := generateLink()
		mu.Lock()
		if _, exist := Links[randomLink]; !exist {
			Links[randomLink] = Link
			mu.Unlock()

			return flagBaseURL + randomLink
		}
	}
}

func GetLink(key string) (string, bool) {
	if value, exist := Links[key]; exist {
		return value, true
	}
	return "", false
}

func GetIddres(c *gin.Context) {
	// Проверяем, что это POST-запрос и Content-Type - text/plain
	path := c.Param("key")
	link, isTrue := GetLink(path)
	if isTrue {
		// Automatically sets the Location header and performs the redirect
		c.Redirect(http.StatusTemporaryRedirect, link)
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}

func AddIddres(c *gin.Context) {
	if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "text/plain") {
		// Чтение тела запроса
		body, err := io.ReadAll(c.Request.Body)
		defer c.Request.Body.Close()
		input := string(body)

		// Проверка на ошибку при чтении или если тело пустое (пустой массив JSON)
		if err != nil || len(body) == 0 {
			c.JSON(http.StatusBadRequest, "Failed to read request body")
			return
		}
		parsedURL, err := url.ParseRequestURI(input)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
			return
		}

		// Если тело содержит пустой массив JSON "[]", также возвращаем ошибку

		// Генерация новой ссылки
		link := AddLink(parsedURL.String())

		// Отправка ответа
		c.String(http.StatusCreated, link)
		return
	}
	c.JSON(http.StatusBadRequest, "Content-Type must be text/plain")
	return
}

func main() {
	parseFlags()
	server := gin.Default()
	server.POST("/", AddIddres)
	server.GET("/:key", GetIddres)
	server.Run(flagRunAddr)
}
