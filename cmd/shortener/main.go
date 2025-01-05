package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var Links = make(map[string]string)

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateLink() string {
	result := ""
	for i := 0; i < 7; i++ {
		indx := rand.Intn(51)
		result += string(charset[indx])
	}
	return result
}

func AddLink(Link string) string {
	for {
		randomLink := generateLink()
		if _, exist := Links[randomLink]; !exist {
			Links[randomLink] = Link
			if flagBaseURL == "" {
				flagBaseURL = "http://localhost:8080/"
			}

			if flagBaseURL[len(flagBaseURL)-1:] != "/" {
				flagBaseURL += "/"
			}
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
		c.JSON(http.StatusBadRequest, nil)
	}
}

func AddIddres(c *gin.Context) {
	if strings.HasPrefix(c.Request.Header.Get("Content-Type"), "text/plain") {
		// Чтение тела запроса
		body, err := io.ReadAll(c.Request.Body)
		defer c.Request.Body.Close()

		// Проверка на ошибку при чтении или если тело пустое (пустой массив JSON)
		if err != nil || len(body) == 0 {
			c.JSON(http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Если тело содержит пустой массив JSON "[]", также возвращаем ошибку
		if string(body) == "[]" {
			c.JSON(http.StatusBadRequest, "Empty array is not allowed")
			return
		}

		// Генерация новой ссылки
		Link := AddLink(string(body))

		// Отправка ответа
		c.String(http.StatusCreated, Link)
	}
}

func main() {
	parseFlags()
	server := gin.Default()
	server.POST("/", AddIddres)
	server.GET("/:key", GetIddres)
	server.Run(flagRunAddr)
}
