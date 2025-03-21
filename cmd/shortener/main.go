package main

import (
	"context"
	"time"

	"github.com/skakunma/go-musthave-shortener-tpl/internal/config"
	"github.com/skakunma/go-musthave-shortener-tpl/internal/handlers"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		panic(err)
	}
	ParseFlags(cfg)
	server := gin.Default()

	handlers.SetupRoutes(server, cfg)

	server.Run(cfg.FlagRunAddr)
	cfg.File.Close()
}
