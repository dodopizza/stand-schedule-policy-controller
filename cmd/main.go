package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/dodopizza/stand-schedule-policy-controller/config"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Logger error: %s", err)
	}
	app.Run(logger, cfg)
}
