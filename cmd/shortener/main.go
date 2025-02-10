package main

import (
	"GoIncrease1/cmd/storage"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	mu            sync.Mutex
	charset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetLength = 7
	sugar         zap.SugaredLogger
	file          *os.File
	store         storage.Storage
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
	gzipResponseWriter struct {
		gin.ResponseWriter
		Writer io.Writer
	}
	shortenTextFile struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	w.responseData.body.Write(b)
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html") {
		return g.Writer.Write(data)
	}
	return g.ResponseWriter.Write(data)
}

func (info *shortenTextFile) SaveURLInfo() error {
	encoder := json.NewEncoder(file)
	err := encoder.Encode(info)
	if err != nil {
		return err
	}
	return nil
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
func AddLink(Link string) (string, error) {
	for {
		randomLink := generateLink()
		mu.Lock()
		if _, exist, _ := store.Get(randomLink); !exist {
			store.Save(randomLink, Link)
			mu.Unlock()
			uuid := strconv.Itoa(store.Len() - 1)
			url := shortenTextFile{UUID: uuid, ShortURL: randomLink, OriginalURL: Link}
			err := url.SaveURLInfo()
			if err != nil {
				return "", err
			}
			return flagBaseURL + randomLink, nil
		}
	}
}

func GetLink(key string) (string, bool) {
	if value, exist, err := store.Get(key); exist && err == nil {
		return value, true
	}
	return "", false
}

func GetIddres(c *gin.Context) {
	path := c.Param("key")
	link, isTrue := GetLink(path)
	if isTrue {
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

func gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode gzip request"})
				c.Abort()
				return
			}
			defer gzipReader.Close()

			c.Request.Body = io.NopCloser(gzipReader)
		}

		if strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") ||
			strings.Contains(c.Writer.Header().Get("Content-Type"), "text/html") {
			c.Writer.Header().Set("Content-Encoding", "gzip")
			gzipWriter := gzip.NewWriter(c.Writer)
			defer gzipWriter.Close()

			c.Writer = &gzipResponseWriter{
				ResponseWriter: c.Writer,
				Writer:         gzipWriter,
			}
		}

		c.Next()
	}
}
func AddIddres(c *gin.Context) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "text/plain") &&
		!strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/x-gzip") {
		c.JSON(http.StatusBadRequest, "Content-Type must be text/plain")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	input := string(body)

	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, "Failed to read request body")
		return
	}
	parsedURL, err := url.ParseRequestURI(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	link, err := AddLink(parsedURL.String())
	if err != nil {
		sugar.Error(err)
	}

	c.String(http.StatusCreated, link)
}

func AddIddresJSON(c *gin.Context) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, "Content-Type must be application/json")
		return
	}
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

	link, err := AddLink(parsedURL.String())

	if err != nil {
		sugar.Error(err)
	}

	_, err = json.Marshal(Response{Result: link})
	if err != nil {
		sugar.Infof("Error: %v", err)
		c.JSON(http.StatusBadGateway, "Problem with service")
	}
	c.JSON(http.StatusCreated, Response{Result: link})
}

func StatusConnDB(c *gin.Context) {
	if _, ok := store.(*storage.PostgresStorage); ok {
		if err := store.Ping(); err != nil {
			sugar.Error(err)
			c.Status(http.StatusInternalServerError)
		}
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}

func loadLinksFromFile() error {
	file, err := os.Open(flagPathToSave)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var link shortenTextFile
		err := json.Unmarshal(scanner.Bytes(), &link)
		if err != nil {
			return fmt.Errorf("failed to parse JSON: %v", err)
		}
		store.Save(link.ShortURL, link.OriginalURL)

	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	return nil
}

func main() {
	parseFlags()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar()
	if flagForDB != "" {
		pgStorage, err := storage.NewPostgresStorage(flagForDB)
		if err != nil {
			sugar.Error(err)
		}
		store = pgStorage
	} else {
		store = storage.NewLinkStorage()
	}
	if err := loadLinksFromFile(); err != nil {
		sugar.Error(err)
	}

	file, err = os.OpenFile(flagPathToSave, os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		sugar.Errorf("failed to open file: %w", err)
	}

	server := gin.Default()
	server.Use(WithLogging())
	server.Use(gzipMiddleware())
	server.POST("/", AddIddres)
	server.GET("/:key", GetIddres)
	server.POST("/api/shorten", AddIddresJSON)
	server.GET("/ping", StatusConnDB)
	server.Run(flagRunAddr)
	file.Close()
}
