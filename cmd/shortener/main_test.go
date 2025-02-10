package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"GoIncrease1/cmd/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/", AddIddres)
	r.GET("/:key", GetIddres)
	r.POST("/api/shorten", AddIddresJSON)
	r.GET("/ping", StatusConnDB)
	return r
}
func TestPostAddress(t *testing.T) {
	flagBaseURL = "http://localhost:8080/"
	store = storage.NewLinkStorage()
	r := setupRouter()

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
	flagBaseURL = "http://localhost:8080/"
	store = storage.NewLinkStorage()
	r := setupRouter()

	// Создание сокращенной ссылки
	postReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://google.com"))
	postReq.Header.Set("Content-Type", "text/plain")
	postRecorder := httptest.NewRecorder()
	r.ServeHTTP(postRecorder, postReq)
	assert.Equal(t, http.StatusCreated, postRecorder.Code)

	createdLink := postRecorder.Body.String()
	shortKey := createdLink[len(flagBaseURL):]

	// Проверка редиректа
	getReq := httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
	getRecorder := httptest.NewRecorder()
	r.ServeHTTP(getRecorder, getReq)

	assert.Equal(t, http.StatusTemporaryRedirect, getRecorder.Code)
	assert.Equal(t, "https://google.com", getRecorder.Header().Get("Location"))
}

func TestGetAddressNotFound(t *testing.T) {
	store = storage.NewLinkStorage()
	r := setupRouter()

	getReq := httptest.NewRequest(http.MethodGet, "/nonexistentkey", nil)
	getRecorder := httptest.NewRecorder()
	r.ServeHTTP(getRecorder, getReq)

	assert.Equal(t, http.StatusNotFound, getRecorder.Code)
}

func TestAddAddressJSON(t *testing.T) {
	flagBaseURL = "http://localhost:8080/"
	store = storage.NewLinkStorage()
	r := setupRouter()

	tests := []struct {
		name        string
		contentType string
		request     string
		want        int
		wantPattern string
	}{
		{"Invalid Content-Type", "text/plain", `{"url": "http://example.com"}`, http.StatusBadRequest, ""},
		{"Invalid JSON format", "application/json", `{invalid_json}`, http.StatusBadRequest, ""},
		{"Missing URL in JSON", "application/json", `{"not_url": "http://example.com"}`, http.StatusBadRequest, ""},
		{"Valid URL", "application/json", `{"url": "http://example.com"}`, http.StatusCreated, `^\{"result":"http://localhost:8080/[a-zA-Z0-9]{7}"\}$`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(tt.request))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.want, w.Code)

			if w.Code == http.StatusCreated && tt.wantPattern != "" {
				var resp struct {
					Result string `json:"result"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				match, _ := regexp.MatchString(tt.wantPattern, resp.Result)
				assert.True(t, match)
			}
		})
	}
}
