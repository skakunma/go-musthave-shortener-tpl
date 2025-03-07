package handlers

import (
	"GoIncrease1/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes регистрирует маршруты в Gin
func SetupRoutes(router *gin.Engine) {
	router.Use(middleware.WithLogging())
	router.Use(middleware.GzipMiddleware())
	router.Use(middleware.AuthMiddleware())
	router.POST("/", AddAddress)
	router.GET("/:key", GetAddress)
	router.POST("/api/shorten", AddAddressJSON)
	router.GET("/ping", StatusConnDB)
	router.POST("/api/shorten/batch", Batch)
	router.GET("/api/user/urls", GetAddressFromUser)

}
