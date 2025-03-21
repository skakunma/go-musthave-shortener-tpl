package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var testRouter *gin.Engine
var testConfig *config.Config

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	testConfig, err = config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	testRouter = gin.Default()
	handlers.SetupRoutes(testRouter, testConfig)

	os.Exit(m.Run())
}

type Response struct {
	Result string `json:"result"`
}

func TestPostAddress(t *testing.T) {
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
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.want, w.Code)
			if w.Code == http.StatusCreated {
				match, _ := regexp.MatchString(`^http://localhost:8080/[a-zA-Z0-9]{7}$`, w.Body.String())
				assert.True(t, match)
			}
		})
	}
}

func TestGetAddressNotFound(t *testing.T) {
	getReq := httptest.NewRequest(http.MethodGet, "/nonexistentkey", nil)
	getRecorder := httptest.NewRecorder()
	testRouter.ServeHTTP(getRecorder, getReq)

	assert.Equal(t, http.StatusNotFound, getRecorder.Code)
}

func TestAddAddressJSON(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		request     string
		want        int
		wantPattern string
	}{
		{"Invalid Content-Type", "text/plain", `{"url": "http://example.com"}`, http.StatusBadRequest, ""},
		{"Invalid JSON format", "application/json", `{invalid_json}`, http.StatusBadRequest, ""},
		{"Valid URL", "application/json", `{"url": "http://example.com"}`, http.StatusCreated, `^\{"result":"http://localhost:8080/[a-zA-Z0-9]{7}"\}$`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.request))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.want, w.Code)
			if w.Code == http.StatusCreated && tt.wantPattern != "" {
				match, _ := regexp.MatchString(tt.wantPattern, w.Body.String())
				assert.True(t, match)
			}
		})
	}
}

func TestBatch(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		request     string
		want        int
		wantPattern string
	}{
		{"Good request", "application/json", `[{"correlation_id":"abc123","original_url":"https://example.com/long-url-1"}]`, http.StatusCreated, `^\[{"correlation_id":"[a-zA-Z0-9]+","short_url":"http://localhost:8080/[a-zA-Z0-9]{7}"}\]$`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.request))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.want, w.Code)
			if w.Code == http.StatusCreated && tt.wantPattern != "" {
				match, _ := regexp.MatchString(tt.wantPattern, w.Body.String())
				assert.True(t, match)
			}
		})
	}
}
