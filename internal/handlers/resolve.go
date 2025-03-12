package handlers

import (
	"errors"
	"net/http"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"
	jwtAuth "github.com/skakunma/go-musthave-shortener-tpl/internal/jwt"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/shortener"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/storage"

	"github.com/gin-gonic/gin"
)

type userURL struct {
	ShortenURL  string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func GetAddress(c *gin.Context, cfg *config.Config) {
	path := c.Param("key")
	ctx := c.Request.Context()
	link, found := shortener.GetLink(ctx, cfg, path)
	if found {
		c.Redirect(http.StatusTemporaryRedirect, link)
	} else {
		c.JSON(http.StatusNotFound, nil)
	}
}

func GetAddressFromUser(c *gin.Context, cfg *config.Config) {
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

	result, err := cfg.Store.GetLinksByUserID(ctx, userClaims.UserID)
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
		response = append(response, userURL{ShortenURL: cfg.FlagBaseURL + key, OriginalURL: value})
	}

	c.JSON(http.StatusOK, response)
}
