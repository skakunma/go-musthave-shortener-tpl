package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/", AddIddres)
	r.GET("/:key", GetIddres)
	r.POST("/api/shorten", AddIddresJSON)
	return r
}

func TestPostIddres(t *testing.T) {
	flagBaseURL = "http://localhost:8080/"
	flagPathToSave = "default.txt"
	Links = NewLinkStorage()
	r := setupRouter()

	tests := []struct {
		name    string
		request string
		want    int
	}{
		{name: "Empty request body", request: "", want: http.StatusBadRequest},
		{name: "Valid URL", request: "https://google.com", want: http.StatusCreated},
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

			if w.Code != tt.want {
				t.Errorf("Expected status code %d, but got %d", tt.want, w.Code)
			}

			if w.Code == http.StatusCreated {
				match, _ := regexp.MatchString(`^http://localhost:8080/[a-zA-Z]{7}$`, w.Body.String())
				if !match {
					t.Errorf("Expected response format to match short URL pattern, but got: %s", w.Body.String())
				}
			}
		})
	}
}

func TestGetIddres(t *testing.T) {
	r := gin.Default()
	flagPathToSave = "default.txt"
	r.GET("/:key", GetIddres) // Обработчик для GET-запросов
	r.POST("/", AddIddres)
	flagBaseURL = "http://localhost:8080/"
	Links = NewLinkStorage()
	r := setupRouter()

	postReq := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytes.NewBufferString("https://google.com")))
	postReq.Header.Set("Content-Type", "text/plain")
	postRecorder := httptest.NewRecorder()
	r.ServeHTTP(postRecorder, postReq)

	if postRecorder.Code != http.StatusCreated {
		t.Fatalf("Expected status code %d, but got %d", http.StatusCreated, postRecorder.Code)
	}

	createdLink := postRecorder.Body.String()
	shortKey := createdLink[len(flagBaseURL):]

	getReq := httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
	getRecorder := httptest.NewRecorder()
	r.ServeHTTP(getRecorder, getReq)

	if getRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, but got %d", http.StatusTemporaryRedirect, getRecorder.Code)
	}

	location := getRecorder.Header().Get("Location")
	if location != "https://google.com" {
		t.Errorf("Expected Location header to be 'https://google.com', but got '%s'", location)
	}
}

func TestGetIddresNotFound(t *testing.T) {
	r := setupRouter()

	getReq := httptest.NewRequest(http.MethodGet, "/nonexistentkey", nil)
	getRecorder := httptest.NewRecorder()
	r.ServeHTTP(getRecorder, getReq)

	if getRecorder.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, getRecorder.Code)
	}
}

func TestAddIddresJSON(t *testing.T) {
	flagBaseURL = "http://localhost:8080/"
	Links = NewLinkStorage()
	r := setupRouter()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugar = *logger.Sugar()

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
			wantPattern: `^{"result":"http://localhost:8080/[a-zA-Z0-9]{7}$"}`,
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

				match, _ := regexp.MatchString(tt.wantPattern, resp.Result)
				if !match {
					t.Errorf("Expected response to match pattern %s, but got: %s", tt.wantPattern, resp.Result)
				}
			}
		})
	}
}
