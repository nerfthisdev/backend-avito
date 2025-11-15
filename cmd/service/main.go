package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nerfthisdev/backend-avito/internal/db"
	httpapi "github.com/nerfthisdev/backend-avito/internal/http"
	"github.com/nerfthisdev/backend-avito/internal/logger"
	"github.com/nerfthisdev/backend-avito/internal/repository"
	"github.com/nerfthisdev/backend-avito/internal/service"
	"go.uber.org/zap"
)

func main() {
	godotenv.Load()

	cfg := logger.Config{Env: "DEV", Level: ""}

	port := os.Getenv("SERVICE_PORT")

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

	defer pool.Close()

	logg.Info("successfully initialized db pool")
	if err := pool.Ping(ctx); err != nil {
		logg.Fatal("failed to ping db", zap.Error(err))
	}

	logg.Info("successfully pinged db")

	// DI

	teamRepo := repository.NewTeamRepo(pool)
	teamSvc := service.NewTeamService(teamRepo)
	teamHandlers := httpapi.NewTeamHandler(teamSvc, logg)

	mux := http.NewServeMux()
	teamHandlers.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		logg.Info("starting http server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Fatal("http server error", zap.Error(err))
		}
	}()

	// Wait for shutdown
	<-ctx.Done()
	logg.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logg.Error("server shutdown error", zap.Error(err))
	}

	logg.Info("bye!")
}
