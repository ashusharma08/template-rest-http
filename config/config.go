package config

import (
	"log/slog"

	"github.com/koding/multiconfig"
)

type Config struct {
	Path string `default:"config.yaml"`
}

func LoadConfig() (*Config, error) {
	cfg := new(Config)
	mc := multiconfig.New()
	err := mc.Load(cfg)
	if err != nil {
		slog.Error("error loading config (flags+ENV)", "error", err)
		return nil, err
	}
	if cfg.Path != "" {
		mc = multiconfig.NewWithPath(cfg.Path)
		err = mc.Load(cfg)
		if err != nil {
			slog.Error("error loading config from path ", cfg.Path, "error", err)
			return nil, err
		}
	}
	err = mc.Validate(cfg)
	if err != nil {
		slog.Error("error validating config ", "error", err)
		return nil, err
	}
	return cfg, nil
}
