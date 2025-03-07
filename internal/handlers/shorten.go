package handlers

import (
	"GoIncrease1/internal/config"
	jwtAuth "GoIncrease1/internal/jwt"
	"GoIncrease1/internal/shortener"
	"GoIncrease1/internal/storage"
	"encoding/json"
	"errors"
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
	if exist == false {
		c.JSON(http.StatusUnauthorized, "You are not autorizate")
	}
	userClaims := claims.(*jwtAuth.Claims)

	ctx := c.Request.Context()
	uuid := strconv.Itoa(config.Cfg.Store.Len(ctx) - 1)
	link, err := shortener.AddLink(ctx, parsedURL.String(), uuid, userClaims.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			_, err = json.Marshal(Response{Result: link})
			if err != nil {
				config.Cfg.Sugar.Infof("Error: %v", err)
				c.JSON(http.StatusBadGateway, "Problem with service")
				return
			}
			c.String(http.StatusConflict, link)
			return
		}
		config.Cfg.Sugar.Error(err)
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
	claims, exist := c.Get("user")
	if exist != true {
		c.JSON(http.StatusUnauthorized, "You are not autorizate")
	}
	userClaims := claims.(*jwtAuth.Claims)

	ctx := c.Request.Context()

	uuid := strconv.Itoa(config.Cfg.Store.Len(ctx))

	link, err := shortener.AddLink(ctx, parsedURL.String(), uuid, userClaims.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			_, err = json.Marshal(Response{Result: link})
			if err != nil {
				config.Cfg.Sugar.Infof("Error: %v", err)
				c.JSON(http.StatusBadGateway, "Problem with service")
				return
			}
			c.JSON(http.StatusConflict, Response{Result: link})
			return
		}
		config.Cfg.Sugar.Error(err)
		c.JSON(http.StatusBadRequest, "Error with add Link to storage")
		return
	}
	c.JSON(http.StatusCreated, Response{Result: link})
}
