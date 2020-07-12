package config

import (
	"github.com/rinx/vald-meta-halodb/internal/config"
)

type GlobalConfig = config.GlobalConfig

// Config represent a application setting data content (config.yaml).
// In K8s environment, this configuration is stored in K8s ConfigMap.
type Data struct {
	config.GlobalConfig `json:",inline" yaml:",inline"`

	// Server represent all server configurations
	Server *config.Servers `json:"server_config" yaml:"server_config"`

	// Observability represent observability configurations
	Observability *config.Observability `json:"observability" yaml:"observability"`
}

func NewConfig(path string) (cfg *Data, err error) {
	err = config.Read(path, &cfg)

	if err != nil {
		return nil, err
	}

	if cfg != nil {
		cfg.Bind()
	}

	if cfg.Server != nil {
		cfg.Server = cfg.Server.Bind()
	}

	if cfg.Observability != nil {
		cfg.Observability = cfg.Observability.Bind()
	}

	return cfg, nil
}
