package main

import (
	"context"
	"runtime"

	"github.com/rinx/vald-meta-halodb/internal/info"
	"github.com/rinx/vald-meta-halodb/internal/log"
	"github.com/rinx/vald-meta-halodb/internal/runner"
	"github.com/rinx/vald-meta-halodb/internal/safety"
	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/config"
	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/usecase"
)

const (
	maxVersion = "v0.0.10"
	minVersion = "v0.0.0"
	name       = "meta halodb"
)

func main() {
	runtime.GOMAXPROCS(1)

	if err := safety.RecoverFunc(func() error {
		return runner.Do(
			context.Background(),
			runner.WithName(name),
			runner.WithVersion(info.Version, maxVersion, minVersion),
			runner.WithConfigLoader(func(path string) (interface{}, *config.GlobalConfig, error) {
				cfg, err := config.NewConfig(path)
				if err != nil {
					return nil, nil, err
				}
				return cfg, &cfg.GlobalConfig, nil
			}),
			runner.WithDaemonInitializer(func(cfg interface{}) (runner.Runner, error) {
				return usecase.New(cfg.(*config.Data))
			}),
		)
	})(); err != nil {
		log.Fatal(err, info.Get())
		return
	}
}
