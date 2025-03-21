package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"
	jwtAuth "github.com/skakunma/go-musthave-shortener-tpl/internal/jwt"
)

func deleteUrls(c *gin.Context, cfg *config.Config) {
	// Проверяем Content-Type
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content type must be application/json"})
		return
	}

	claims, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusUnauthorized, "You are not autorizate")
		return
	}
	userClaims := claims.(*jwtAuth.Claims)
	userID := userClaims.UserID

	// Читаем тело запроса
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Декодируем JSON
	var linksUUID []string
	if err := json.Unmarshal(body, &linksUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error with JSON unmarshal"})
		return
	}
	ctx := c.Request.Context()

	var wg sync.WaitGroup

	for _, uuid := range linksUUID {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			author, _ := cfg.Store.GetUserFromUUID(ctx, id)
			if author == userID {
				cfg.Store.DeleteURL(ctx, id)
			}
		}(uuid)
	}

	wg.Wait()

	c.JSON(http.StatusAccepted, gin.H{"message": "URLs deleted successfully"})
}
