package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var Links = make(map[string]string)

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateLink() string {
	result := ""
	for i := 0; i < 7; i++ {
		indx := rand.Intn(51)
		result += string(charset[indx])
	}
	return result
}

func AddLink(Link string) string {
	for {
		randomLink := generateLink()
		if _, exist := Links[randomLink]; !exist {
			Links[randomLink] = Link
			return "http://localhost:8080/" + randomLink
		}
	}
}

func GetLink(key string) (string, bool) {
	if value, exist := Links[key]; exist {
		return value, true
	}
	return "", false
}

func Iddres(res http.ResponseWriter, req *http.Request) {
	// Проверяем, что это POST-запрос и Content-Type - text/plain
	if req.Method == http.MethodPost && strings.HasPrefix(req.Header.Get("Content-Type"), "text/plain") {
		// Чтение тела запроса
		body, err := io.ReadAll(req.Body)
		defer req.Body.Close() // Обязательно закрыть тело запроса после чтения

		// Проверка на ошибку при чтении или если тело пустое (пустой массив JSON)
		if err != nil || len(body) == 0 {
			http.Error(res, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Если тело содержит пустой массив JSON "[]", также возвращаем ошибку
		if string(body) == "[]" {
			http.Error(res, "Empty array is not allowed", http.StatusBadRequest)
			return
		}

		// Генерация новой ссылки
		Link := AddLink(string(body))

		// Отправка ответа
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(Link))
	}

	if req.Method == http.MethodGet {
		path := strings.Trim(req.URL.Path, "/")
		link, isTrue := GetLink(path)
		if isTrue {
			// Automatically sets the Location header and performs the redirect
			http.Redirect(res, req, link, http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
	}
}

func main() {
	mx := http.NewServeMux()
	mx.HandleFunc("/", Iddres)
	if err := http.ListenAndServe(":8080", mx); err != nil {
		panic(err)
	}
}
