package handlers

import (
	"GoIncrease1/internal/config"
	jwtAuth "GoIncrease1/internal/jwt"
	"GoIncrease1/internal/shortener"
	"GoIncrease1/internal/storage"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type userURL struct {
	ShortenURL  string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func GetAddress(c *gin.Context) {
	path := c.Param("key")
	ctx := c.Request.Context()
	link, found := shortener.GetLink(ctx, path)
	if found {
		c.Redirect(http.StatusTemporaryRedirect, link)
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}

func GetAddressFromUser(c *gin.Context) {
	claims, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusUnauthorized, "You are not authorized")
		return
	}

	userClaims, ok := claims.(*jwtAuth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, "Invalid token data")
		return
	}

	ctx := c.Request.Context()

	result, err := config.Cfg.Store.GetLinksByUserId(ctx, userClaims.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			c.JSON(http.StatusNoContent, []userURL{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNoContent, []userURL{})
		return
	}

	response := make([]userURL, 0, len(result))
	for key, value := range result {
		response = append(response, userURL{ShortenURL: config.Cfg.FlagBaseURL + key, OriginalURL: value})
	}

	c.JSON(http.StatusOK, response)
}
