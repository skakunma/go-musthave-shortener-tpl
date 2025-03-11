package handlers

import (
	"net/http"
	"strings"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"
	jwtAuth "github.com/skakunma/go-musthave-shortener-tpl/internal/jwt"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/shortener"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/storage"

	"github.com/gin-gonic/gin"
)

type infoAboutURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	ShortLink     string
}

type infoAboutURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortLink     string `json:"short_url"`
}

func Batch(c *gin.Context, cfg *config.Config) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type must be application/json"})
		return
	}

	var links []storage.InfoAboutURL
	if err := c.ShouldBindJSON(&links); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	ctx := c.Request.Context()
	claims, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not authorized"})
		return
	}
	userClaims := claims.(*jwtAuth.Claims)

	for i, link := range links {
		if link.OriginalURL == "" || link.CorrelationID == "" {
			c.JSON(http.StatusBadRequest, "JSON is not correctly")
			return
		}
		links[i].ShortLink = shortener.GenerateLink(cfg)
	}

	_, err := cfg.Store.AddLinksBatch(ctx, links, userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Problem service")
		return
	}
	var response []infoAboutURLResponse
	for _, link := range links {
		response = append(response, infoAboutURLResponse{
			CorrelationID: link.CorrelationID,
			ShortLink:     cfg.FlagBaseURL + link.ShortLink,
		})
	}

	c.JSON(http.StatusCreated, response)
}
