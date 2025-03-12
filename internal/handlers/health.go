package handlers

import (
	"net/http"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"

	"github.com/gin-gonic/gin"
)

func StatusConnDB(c *gin.Context, cfg *config.Config) {
	ctx := c.Request.Context()
	if err := cfg.Store.Ping(ctx); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
