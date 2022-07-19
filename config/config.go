package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/http"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
)

type (
	Config struct {
		Kube  kubernetes.Config `json:"kube"`
		Azure azure.Config      `json:"azure"`
		Http  http.Config       `json:"http"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadConfig("./config/config.json", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, cleanenv.ReadEnv(cfg)
}
