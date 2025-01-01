package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestPostIddres(t *testing.T) {
	type want struct {
		Response string
		Code     int
	}
	tests := []struct {
		name    string
		want    want
		request string
	}{
		{
			name: "Testing 400",
			want: want{
				Code:     400,
				Response: "Failed to read request body\n",
			},
			request: "",
		},
		{
			name: "Testing 201",
			want: want{
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
			assert.Equal(t, w.Code, test.want.Code, "The code does not match what was expected")
			response := w.Body.String()
			if test.want.Code == 400 {
				assert.Equal(t, response, test.want.Response, "Answer is incorrect")
			} else {
				match, _ := regexp.MatchString(`^http://localhost:8080/[a-zA-Z]{7}$`, response)
				assert.True(t, match, "Response URL format is incorrect")
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
	assert.Equal(t, getRecorder.Code, http.StatusTemporaryRedirect, "Expected redirect status")
	assert.Equal(t, getRecorder.Header().Get("Location"), "https://google.com", "Expected correct redirection")
}

func TestGetIddresNotFound(t *testing.T) {
	// Выполняем GET-запрос для несуществующего ключа
	getReq := httptest.NewRequest(http.MethodGet, "/nonexistentkey", nil)
	getRecorder := httptest.NewRecorder()
	Iddres(getRecorder, getReq)

	// Проверяем, что мы получили ошибку 400
	assert.Equal(t, getRecorder.Code, http.StatusBadRequest, "Expected 400 status for nonexistent key")
}
