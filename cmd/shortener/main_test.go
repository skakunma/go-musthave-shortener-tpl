package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestPostIddres(t *testing.T) {
	tests := []struct {
		name string
		want struct {
			Response string
			Code     int
		}
		request string
	}{
		{
			name: "Testing 400",
			want: struct {
				Response string
				Code     int
			}{
				Code:     400,
				Response: "Failed to read request body\n",
			},
			request: "",
		},
		{
			name: "Testing 201",
			want: struct {
				Response string
				Code     int
			}{
				Code:     201,
				Response: "",
			},
			request: "https://google.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var req *http.Request
			if test.request == "" {
				req = httptest.NewRequest(http.MethodPost, "/", nil)
			} else {
				req = httptest.NewRequest(http.MethodPost, "/", io.NopCloser(io.Reader(bytes.NewBufferString(test.request))))
			}
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			Iddres(w, req)

			// Проверка кода статуса
			if w.Code != test.want.Code {
				t.Errorf("Expected status code %d, but got %d", test.want.Code, w.Code)
			}

			// Проверка тела ответа
			response := w.Body.String()
			if test.want.Code == 400 {
				if response != test.want.Response {
					t.Errorf("Expected response body '%s', but got '%s'", test.want.Response, response)
				}
			} else {
				match, _ := regexp.MatchString(`^http://localhost:8080/[a-zA-Z]{7}$`, response)
				if !match {
					t.Errorf("Expected response URL format to match '^http://localhost:8080/[a-zA-Z]{7}$', but got '%s'", response)
				}
			}
		})
	}
}

func TestGetIddres(t *testing.T) {
	// Сначала создаем ссылку через POST-запрос
	postReq := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(io.Reader(bytes.NewBufferString("https://google.com"))))
	postReq.Header.Set("Content-Type", "text/plain")
	postRecorder := httptest.NewRecorder()
	Iddres(postRecorder, postReq)

	// Проверяем, что ссылка была успешно создана
	createdLink := postRecorder.Body.String()

	// Теперь выполняем GET-запрос по этой ссылке
	getReq := httptest.NewRequest(http.MethodGet, createdLink, nil)
	getRecorder := httptest.NewRecorder()
	Iddres(getRecorder, getReq)

	// Проверяем, что мы получили редирект
	if getRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected redirect status %d, but got %d", http.StatusTemporaryRedirect, getRecorder.Code)
	}

	location := getRecorder.Header().Get("Location")
	if location != "https://google.com" {
		t.Errorf("Expected Location header to be 'https://google.com', but got '%s'", location)
	}
}

func TestGetIddresNotFound(t *testing.T) {
	// Выполняем GET-запрос для несуществующего ключа
	getReq := httptest.NewRequest(http.MethodGet, "/nonexistentkey", nil)
	getRecorder := httptest.NewRecorder()
	Iddres(getRecorder, getReq)

	// Проверяем, что мы получили ошибку 400
	if getRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, getRecorder.Code)
	}
}
