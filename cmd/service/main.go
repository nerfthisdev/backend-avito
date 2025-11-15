package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nerfthisdev/backend-avito/internal/db"
	"github.com/nerfthisdev/backend-avito/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := logger.Config{Env: "DEV", Level: ""}

	logg, err := logger.New(cfg)
	if err != nil {
		panic("couldnt initialize zap logger")
	}

	logg.Info("successfully initialized zap logger")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		sig := <-ch
		logg.Info("received signal, shutting down...", zap.String("signal", sig.String()))
		cancel()
	}()

	dbCfg := db.FromEnv()
	pool, err := db.NewPool(ctx, dbCfg)
	if err != nil {
		logg.Fatal("failed to init db", zap.Error(err))
	}

	logg.Info("successfully initialized db pool")
	if err := pool.Ping(ctx); err != nil {
		logg.Fatal("failed to ping db", zap.Error(err))
	}

	logg.Info("successfully pinged db")

	defer pool.Close()
}
