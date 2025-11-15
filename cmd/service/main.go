package main

import "github.com/nerfthisdev/backend-avito/internal/logger"

func main() {
	cfg := logger.Config{Env: "DEV", Level: ""}

	logg, err := logger.New(cfg)
	if err != nil {
		panic("couldnt initialize logger")
	}

	logg.Info("successfully initialized logger")
}
