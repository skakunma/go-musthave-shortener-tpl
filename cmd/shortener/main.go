package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"go.uber.org/zap"
	"bytes"

)

var (
	mu      sync.Mutex
	Links   = make(map[string]string)
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetLength = 7
    sugar zap.SugaredLogger

)

type responseData struct {
	status int
	size   int
	body   *bytes.Buffer
}

// Обертка для ResponseWriter
type loggingResponseWriter struct {
	gin.ResponseWriter
	responseData *responseData
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	// Записываем данные в буфер
	w.responseData.body.Write(b)
	// Записываем данные в оригинальный ResponseWriter
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}


func generateLink() string {
	var builder strings.Builder
	builder.Grow(charsetLength)

	for i := 0; i < charsetLength; i++ {
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

func WithLogging() gin.HandlerFunc {
	logFn := func(c *gin.Context) {
		start := time.Now()
		uri := c.Request.RequestURI
		method := c.Request.Method
		responseData := &responseData {
			status: 0,
			size: 0,
			body:   new(bytes.Buffer),
		}
		lw := &loggingResponseWriter{
			ResponseWriter: c.Writer,
			responseData:   responseData,
		}
		c.Writer = lw
		c.Next()
		duration := time.Since(start)
		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"response_size", responseData.size,
			"response_status", responseData.status,

			)
	}
	return gin.HandlerFunc(logFn)
}

func AddIddres(c *gin.Context) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "text/plain"){
		c.JSON(http.StatusBadRequest, "Content-Type must be text/plain")
		return
	}
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
}

func main() {
	parseFlags()
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar()
	server := gin.Default()
	server.Use(WithLogging())
	server.POST("/", AddIddres)
	server.GET("/:key", GetIddres)
	server.Run(flagRunAddr)
}
//For start Actions