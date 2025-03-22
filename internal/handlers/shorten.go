package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"encoding/hex"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"
	jwtAuth "github.com/skakunma/go-musthave-shortener-tpl/internal/jwt"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/shortener"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/storage"

	"github.com/gin-gonic/gin"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func GenerateUUID(shorten string) string {
	hash := sha256.Sum256([]byte(shorten))
	return hex.EncodeToString(hash[:])
}

func AddAddress(c *gin.Context, cfg *config.Config) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "text/plain") &&
		!strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/x-gzip") {
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
	claims, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusUnauthorized, "You are not autorizate")
	}
	userClaims := claims.(*jwtAuth.Claims)

	ctx := c.Request.Context()
	uuid := GenerateUUID(parsedURL.String())
	link, err := shortener.AddLink(ctx, cfg, parsedURL.String(), uuid, userClaims.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			c.String(http.StatusConflict, link)
			return
		}
		cfg.Sugar.Error(err)
		c.String(http.StatusBadRequest, "Error with add Link to storage")
		return
	}
	c.String(http.StatusCreated, link)

}

func AddAddressJSON(c *gin.Context, cfg *config.Config) {
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
	claims, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusUnauthorized, "You are not autorizate")
	}
	userClaims := claims.(*jwtAuth.Claims)

	ctx := c.Request.Context()

	uuid := GenerateUUID(parsedURL.String())

	link, err := shortener.AddLink(ctx, cfg, parsedURL.String(), uuid, userClaims.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			_, err = json.Marshal(Response{Result: link})
			if err != nil {
				cfg.Sugar.Infof("Error: %v", err)
				c.JSON(http.StatusBadGateway, "Problem with service")
				return
			}
			c.JSON(http.StatusConflict, Response{Result: link})
			return
		}
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadRequest, "Error with add Link to storage")
		return
	}
	c.JSON(http.StatusCreated, Response{Result: link})
}
