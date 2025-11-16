package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg Config) (*zap.Logger, error) {
	var zapCfg zap.Config

	switch strings.ToLower(cfg.Env) {
	case "dev", "development":
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.Encoding = "console"
	default:
		zapCfg = zap.NewProductionConfig()
		zapCfg.Encoding = "json"
	}

	if cfg.Level != "" {
		var lvl zapcore.Level
		if err := lvl.UnmarshalText([]byte(cfg.Level)); err == nil {
			zapCfg.Level = zap.NewAtomicLevelAt(lvl)
		}
	}

	return zapCfg.Build()
}
