package middleware

import (
	"GoIncrease1/internal/config"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

		if strings.Contains(c.GetHeader("Content-Type"), "application/json") ||
			strings.Contains(c.GetHeader("Content-Type"), "text/html") {
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
