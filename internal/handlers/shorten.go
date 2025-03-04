package handlers

import (
	"GoIncrease1/internal/config"
	"GoIncrease1/internal/shortener"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func AddAddress(c *gin.Context) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "text/plain") {
		c.JSON(http.StatusBadRequest, "Content-Type must be text/plain")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, "Failed to read request body")
		return
	}
	parsedURL, err := url.ParseRequestURI(string(body))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	uuid := strconv.Itoa(config.Cfg.Store.Len() - 1)
	link, err := shortener.AddLink(parsedURL.String(), uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error saving link")
		return
	}
	c.String(http.StatusCreated, link)
}

func AddAddressJSON(c *gin.Context) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	var input Request
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON")
		return
	}

	parsedURL, err := url.ParseRequestURI(input.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	uuid := strconv.Itoa(config.Cfg.Store.Len())
	link, err := shortener.AddLink(parsedURL.String(), uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error saving link")
		return
	}
	c.JSON(http.StatusCreated, Response{Result: link})
}
