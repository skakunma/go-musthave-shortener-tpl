package handlers

import (
	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes регистрирует маршруты в Gin
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	router.Use(middleware.WithLogging(cfg))
	router.Use(middleware.GzipMiddleware())
	router.Use(middleware.AuthMiddleware(cfg))

	// Передаем cfg в обработчики
	router.POST("/", func(c *gin.Context) { AddAddress(c, cfg) })
	router.GET("/:key", func(c *gin.Context) { GetAddress(c, cfg) })
	router.POST("/api/shorten", func(c *gin.Context) { AddAddressJSON(c, cfg) })
	router.GET("/ping", func(c *gin.Context) { StatusConnDB(c, cfg) })
	router.POST("/api/shorten/batch", func(c *gin.Context) { Batch(c, cfg) })
	router.GET("/api/user/urls", func(c *gin.Context) { GetAddressFromUser(c, cfg) })
	router.DELETE("/api/user/urls", func(c *gin.Context) { deleteUrls(c, cfg) })
}
