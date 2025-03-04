package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"go.uber.org/zap"

	"GoIncrease1/internal/config"
	"GoIncrease1/internal/handlers"
	"GoIncrease1/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPostAddress(t *testing.T) {
	config.Cfg = config.NewConfig()
	config.Cfg.FlagPathToSave = "default.txt"
	config.Cfg.File, _ = os.OpenFile(config.Cfg.FlagPathToSave, os.O_CREATE|os.O_RDWR, 0644)
	config.Cfg.FlagBaseURL = "http://localhost:8080/"
	config.Cfg.Store = storage.NewLinkStorage()
	r := gin.Default()
	handlers.SetupRoutes(r)

	tests := []struct {
		name    string
		request string
		want    int
	}{
		{"Empty request body", "", http.StatusBadRequest},
		{"Valid URL", "https://google.com", http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.request == "" {
				req = httptest.NewRequest(http.MethodPost, "/", nil)
			} else {
				req = httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytes.NewBufferString(tt.request)))
			}
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.want, w.Code)

			if w.Code == http.StatusCreated {
				match, _ := regexp.MatchString(`^http://localhost:8080/[a-zA-Z0-9]{7}$`, w.Body.String())
				assert.True(t, match)
			}
		})
	}
}

func TestGetAddress(t *testing.T) {
	config.Cfg = config.NewConfig()
	config.Cfg.FlagPathToSave = "default.txt"
	config.Cfg.File, _ = os.OpenFile(config.Cfg.FlagPathToSave, os.O_CREATE|os.O_RDWR, 0644)
	config.Cfg.FlagBaseURL = "http://localhost:8080/"
	config.Cfg.Store = storage.NewLinkStorage()
	r := gin.Default()
	handlers.SetupRoutes(r)
	// Создание сокращенной ссылки
	postReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://google.com"))
	postReq.Header.Set("Content-Type", "text/plain")
	postRecorder := httptest.NewRecorder()
	r.ServeHTTP(postRecorder, postReq)

	assert.Equal(t, http.StatusCreated, postRecorder.Code)

	createdLink := postRecorder.Body.String()
	shortKey := createdLink[len(config.Cfg.FlagBaseURL):]

	// Проверка редиректа
	getReq := httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
	getRecorder := httptest.NewRecorder()
	r.ServeHTTP(getRecorder, getReq)

	assert.Equal(t, http.StatusTemporaryRedirect, getRecorder.Code)
	assert.Equal(t, "https://google.com", getRecorder.Header().Get("Location"))
}

func TestGetAddressNotFound(t *testing.T) {
	config.Cfg = config.NewConfig()

	config.Cfg.Store = storage.NewLinkStorage()
	r := gin.Default()
	handlers.SetupRoutes(r)

	getReq := httptest.NewRequest(http.MethodGet, "/nonexistentkey", nil)
	getRecorder := httptest.NewRecorder()
	r.ServeHTTP(getRecorder, getReq)

	assert.Equal(t, http.StatusNotFound, getRecorder.Code)
}

func TestAddIddresJSON(t *testing.T) {
	config.Cfg = config.NewConfig()
	config.Cfg.FlagPathToSave = "default.txt"

	config.Cfg.File, _ = os.OpenFile(config.Cfg.FlagPathToSave, os.O_CREATE|os.O_RDWR, 0644)
	defer config.Cfg.File.Close()
	config.Cfg.FlagBaseURL = "http://localhost:8080/"
	config.Cfg.Store = storage.NewLinkStorage()
	r := gin.Default()
	handlers.SetupRoutes(r)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	config.Cfg.Sugar = logger.Sugar()

	tests := []struct {
		name        string
		contentType string
		request     string
		want        int
		wantPattern string
	}{
		{
			name:        "Invalid Content-Type",
			contentType: "text/plain",
			request:     `{"url": "http://example.com"}`,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Invalid JSON format",
			contentType: "application/json",
			request:     `{invalid_json}`,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Missing URL in JSON",
			contentType: "application/json",
			request:     `{"not_url": "http://example.com"}`,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Valid URL",
			contentType: "application/json",
			request:     `{"url": "http://example.com"}`,
			want:        http.StatusCreated,
			wantPattern: `^{"result":"http://localhost:8080/[a-zA-Z0-9]{7}"}$`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.request))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("Expected status code %d, but got %d", tt.want, w.Code)
			}

			if w.Code == http.StatusCreated && tt.wantPattern != "" {
				var resp Response
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				if err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
				}

				match, _ := regexp.MatchString(tt.wantPattern, w.Body.String())
				if !match {
					t.Errorf("Expected response to match pattern %s, but got: %s", tt.wantPattern, w.Body.String())
				}
			}
		})
	}

}

func TestBath(t *testing.T) {
	config.Cfg = config.NewConfig()

	file, err := os.OpenFile(config.Cfg.FlagPathToSave, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	config.Cfg.FlagBaseURL = "http://localhost:8080/"
	config.Cfg.FlagForDB = "host=localhost user=postgres password=example dbname=postgres sslmode=disable"
	config.Cfg.Store = storage.NewLinkStorage()
	r := gin.Default()
	handlers.SetupRoutes(r)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	config.Cfg.Sugar = logger.Sugar()
	tests := []struct {
		name        string
		contentType string
		request     string
		want        int
		wantPattern string
	}{
		{
			name:        "Invalid Content-Type",
			contentType: "text/plain",
			request: `[
					{
						"correlation_id":"a1b2c3d4e5"
						"original_url": "https://www.google.com/"
					}
				]`,
			want: http.StatusBadRequest,
		},
		{
			name:        "invalid JSON",
			contentType: "application/json",
			request:     `{not valid}`,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Missing URL in JSON",
			contentType: "application/json",
			request:     `[{"not_url": "http://example.com"}]`,
			want:        http.StatusBadRequest,
		},
		{
			name:        "Good request",
			contentType: "application/json",
			request: `[
    	{
        "correlation_id": "abc123",
        "original_url": "https://example.com/long-url-1"
    	},
    	{
        "correlation_id": "def456",
        "original_url": "https://example.com/long-url-2"
    	}
	]`,
			want:        http.StatusCreated,
			wantPattern: `^\[{"correlation_id":"[a-zA-Z0-9\-]+","short_url":"http://localhost:8080/[a-zA-Z0-9]{7}"}(,{"correlation_id":"[a-zA-Z0-9\-]+","short_url":"http://localhost:8080/[a-zA-Z0-9]{7}"})*\]$`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.request))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Errorf("Expected status code %d, but got %d", tt.want, w.Code)
			}
			if w.Code == http.StatusCreated && tt.wantPattern != "" {
				if err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
				}

				match, _ := regexp.MatchString(tt.wantPattern, w.Body.String())
				if !match {
					t.Errorf("Expected response to match pattern %s, but got: %s", tt.wantPattern, w.Body.String())
				}
			}
		})
	}
}
