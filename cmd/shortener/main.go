package main

import (
	"GoIncrease1/internal/config"
	"GoIncrease1/internal/handlers"
	"GoIncrease1/internal/storage"
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type (
	Request struct {
		URL string `json:"url"`
	}
	Response struct {
		Result string `json:"result"`
	}
)

func main() {
	config.Cfg = config.NewConfig()
	parseFlags()
	logger, err := zap.NewDevelopment()
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer logger.Sync()
	config.Cfg.Sugar = logger.Sugar()
	if config.Cfg.FlagForDB != "" {
		pgStorage, err := storage.NewPostgresStorage(config.Cfg.FlagForDB)
		if err != nil {
			config.Cfg.Sugar.Error(err)
		}
		config.Cfg.Store = pgStorage
	} else {
		config.Cfg.Store = storage.NewLinkStorage()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := config.LoadLinksFromFile(ctx); err != nil {
		config.Cfg.Sugar.Error(err)
	}

	config.Cfg.File, err = os.OpenFile(config.Cfg.FlagPathToSave, os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		config.Cfg.Sugar.Errorf("failed to open file: %w", err)
	}

	server := gin.Default()

	handlers.SetupRoutes(server)

	server.Run(config.Cfg.FlagRunAddr)
	config.Cfg.File.Close()
}
