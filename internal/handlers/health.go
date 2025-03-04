package handlers

import (
	"GoIncrease1/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StatusConnDB(c *gin.Context) {
	if err := config.Cfg.Store.Ping(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
