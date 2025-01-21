package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	mu            sync.Mutex
	Links         = make(map[string]string)
	charset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetLength = 7
	sugar         zap.SugaredLogger
)

type (
	responseData struct {
		status int
		size   int
		body   *bytes.Buffer
	}
	loggingResponseWriter struct {
		gin.ResponseWriter
		responseData *responseData
	}
	Request struct {
		URL string `json:"url"`
	}
	Response struct {
		Result string `json:"result"`
	}
)

// Обертка для ResponseWriter
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
		responseData := &responseData{
			status: 0,
			size:   0,
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
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "text/plain") {
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

func AddIddresJSON(c *gin.Context) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, "Content-Type must be application/json")
		return
	}
	// Чтение тела запроса
	var (
		Inputurl Request
		buf      bytes.Buffer
	)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		sugar.Infoln("Probblem with serilizator")
		c.JSON(http.StatusBadRequest, "In body must be JSON like this")
		return

	}

	if err = json.Unmarshal(buf.Bytes(), &Inputurl); err != nil {
		sugar.Infoln("Probblem with serilizator")
		c.JSON(http.StatusBadRequest, "In body must be JSON like this")
		return

	}
	defer c.Request.Body.Close()
	body := Inputurl.URL
	parsedURL, err := url.ParseRequestURI(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	// Если тело содержит пустой массив JSON "[]", также возвращаем ошибку

	link := AddLink(parsedURL.String())
	_, err = json.Marshal(Response{Result: link})
	if err != nil {
		sugar.Infof("Error: %v", err)
		c.JSON(http.StatusBadGateway, "Problem with service")
	}
	// Отправка ответа
	c.JSON(http.StatusCreated, Response{Result: link})
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
	server.POST("/api/shorten", AddIddresJSON)
	server.Run(flagRunAddr)
}

//For start Actions
