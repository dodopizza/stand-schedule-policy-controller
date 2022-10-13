package main

import (
	"log"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"

	"github.com/dodopizza/stand-schedule-policy-controller/config"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	app.Run(setupLoggers(), cfg)
}

func setupLoggers() *zap.Logger {
	// replace klog logging with dedicated zap & set upper verbosity
	klog.SetLogger(
		zapr.NewLogger(
			createLogger(zapcore.WarnLevel),
		),
	)

	return createLogger(zapcore.DebugLevel)
}

func createLogger(lvl zapcore.Level) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Logger error: %v", err)
	}

	return logger
}
