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
	for i := 0; i < 5; i++ {
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
			return "http://localhost:8000/" + randomLink
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
	if req.Method == http.MethodPost && req.Header.Get("Content-Type") == "text/plain" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			panic(err)
		}
		fmt.Println(string(body))
		Link := AddLink(string(body))
		res.WriteHeader(http.StatusCreated)
		fmt.Fprintln(res, Link)
	}
	if req.Method == http.MethodGet {
		path := strings.Trim(req.URL.Path, "/")
		fmt.Println(path)
		link, isTrue := GetLink(path)
		if isTrue {
			res.Header().Set("Location", link)
			fmt.Println(res.Header().Get("Location"))
			http.Redirect(res, req, link, http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(res, nil)
		}
	}

}

func main() {
	mx := http.NewServeMux()
	mx.HandleFunc("/", AddIddres)
	if err := http.ListenAndServe(":8000", mx); err != nil {
		panic(err)
	}
}
