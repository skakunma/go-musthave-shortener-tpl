package main

import (
	"fmt"
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

func AddIddres(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost && strings.HasPrefix(req.Header.Get("Content-Type"), "text/plain") {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			panic(err)
		}
		Link := AddLink(string(body))
		res.WriteHeader(http.StatusCreated) // Correct status code
		fmt.Fprintln(res, Link)
	}

	if req.Method == http.MethodGet {
		path := strings.Trim(req.URL.Path, "/")
		link, isTrue := GetLink(path)
		if isTrue {
			// Automatically sets the Location header and performs the redirect
			http.Redirect(res, req, link, http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(res, "Invalid link")
		}
	}
}

func main() {
	mx := http.NewServeMux()
	mx.HandleFunc("/", AddIddres)
	if err := http.ListenAndServe(":8080", mx); err != nil {
		panic(err)
	}
}
