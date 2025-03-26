package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"
	jwtauth "github.com/skakunma/go-musthave-shortener-tpl/internal/jwt"
)

func StartDeleteWorker(cfg *config.Config) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var batch []string
		Loop:
			for {
				select {
				case uuid := <-cfg.DeleteQueue:
					batch = append(batch, uuid)
				default:
					break Loop
				}
			}
			if len(batch) == 0 {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			for _, uuid := range batch {
				if err := cfg.Store.DeleteURL(ctx, uuid); err != nil {
					cfg.Sugar.Error(fmt.Sprintf("Error deleting URL %s: %v", uuid, err))
				}
			}
			cancel()
		}
	}()
}

func deleteUrls(c *gin.Context, cfg *config.Config) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content type must be application/json"})
		return
	}

	userID, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claims := userID.(*jwtauth.Claims)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var uuids []string
	if err := json.Unmarshal(body, &uuids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	var wg sync.WaitGroup
	for _, uuid := range uuids {
		wg.Add(1)
		go func(uuid string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			defer wg.Done()

			owner, err := cfg.Store.GetUserFromUUID(ctx, uuid)
			if err != nil {
				cfg.Sugar.Errorf("Error getting user for %s: %v", uuid, err)
				return
			}
			if owner == claims.UserID {
				cfg.DeleteQueue <- uuid
			}
		}(uuid)
	}
	wg.Wait()

	c.JSON(http.StatusAccepted, gin.H{"message": "URLs scheduled for deletion"})

}
