package middleware

import (
	"GoIncrease1/internal/config"
	jwtAuth "GoIncrease1/internal/jwt"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
	gzipResponseWriter struct {
		gin.ResponseWriter
		Writer io.Writer
	}
)

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
		config.Cfg.Sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"response_size", responseData.size,
			"response_status", responseData.status,
		)
	}
	return gin.HandlerFunc(logFn)
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Обработка входящего запроса с сжатием gzip
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			// Создаем gzip-ридер для распаковки сжатого тела запроса
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode gzip request"})
				c.Abort()
				return
			}
			defer gzipReader.Close()

			// Перенаправляем тело запроса в распакованное содержимое
			c.Request.Body = io.NopCloser(gzipReader)
		}

		// Обработка ответов с сжатием для клиентов, поддерживающих gzip
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") &&
			(strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") ||
				strings.Contains(c.Writer.Header().Get("Content-Type"), "text/html")) {

			// Устанавливаем Content-Encoding в gzip
			c.Writer.Header().Set("Content-Encoding", "gzip")

			// Создаем gzip-редактор для сжатия данных в ответе
			gzipWriter := gzip.NewWriter(c.Writer)
			defer gzipWriter.Close()

			// Переопределяем c.Writer на новый gzip-ответ
			c.Writer = &gzipResponseWriter{
				ResponseWriter: c.Writer,
				Writer:         gzipWriter,
			}
		}

		// Дальше передаем управление следующему middleware или обработчику
		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtToken, err := c.Cookie("jwt")
		if err != nil || jwtToken == "" {
			cx := c.Request.Context()
			newUser, err := config.Cfg.Store.GetNewUser(cx)
			if err != nil {
				config.Cfg.Sugar.Error("Ошибка создания нового пользователя:", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
				return
			}

			token, err := jwtAuth.BuildJWTString(newUser)
			if err != nil {
				config.Cfg.Sugar.Error("Ошибка генерации JWT:", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
				return
			}

			err = config.Cfg.Store.SaveUser(cx, newUser)

			if err != nil {
				config.Cfg.Sugar.Error("Ошибка сохранения пользовалтеля", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
				return
			}

			jwtToken = token
		}

		claims := &jwtAuth.Claims{}
		token, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtAuth.SECRET_KEY), nil
		})
		if err != nil {
			config.Cfg.Sugar.Error("Ошибка парсинга JWT:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Невалидный токен"})
			return
		}
		c.SetCookie("jwt", jwtToken, 3600, "/", "", false, false)
		c.Set("user", claims)
		c.Next()
	}
}
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
