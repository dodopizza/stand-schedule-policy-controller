package controller

import (
	"time"
)

type (
	Config struct {
		ResyncSeconds int `json:"resync_seconds" env:"CONTROLLER_RESYNC_SECONDS"`
	}
)

const (
	_DefaultResyncSeconds = 300 // 5 min
)

func (c *Config) GetResyncDuration() time.Duration {
	if c.ResyncSeconds < _DefaultResyncSeconds {
		return time.Duration(_DefaultResyncSeconds) * time.Second
	}
	return time.Duration(c.ResyncSeconds) * time.Second
}
