package usecase

import (
	"context"

	iconf "github.com/rinx/vald-meta-halodb/internal/config"
	"github.com/rinx/vald-meta-halodb/internal/errgroup"
	"github.com/rinx/vald-meta-halodb/internal/net/grpc"
	"github.com/rinx/vald-meta-halodb/internal/net/grpc/metric"
	"github.com/rinx/vald-meta-halodb/internal/observability"
	"github.com/rinx/vald-meta-halodb/internal/runner"
	"github.com/rinx/vald-meta-halodb/internal/safety"
	"github.com/rinx/vald-meta-halodb/internal/servers/server"
	"github.com/rinx/vald-meta-halodb/internal/servers/starter"
	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/config"
	handler "github.com/rinx/vald-meta-halodb/pkg/meta/halodb/handler/grpc"
	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/handler/rest"
	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/router"
	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/service"
	"github.com/vdaas/vald/apis/grpc/meta"
)

type run struct {
	eg            errgroup.Group
	cfg           *config.Data
	h             service.HaloDB
	server        starter.Server
	observability observability.Observability
}

func New(cfg *config.Data) (r runner.Runner, err error) {
	h, err := service.New()
	if err != nil {
		return nil, err
	}
	g := handler.New(handler.WithHaloDB(h))
	eg := errgroup.Get()

	grpcServerOptions := []server.Option{
		server.WithGRPCRegistFunc(func(srv *grpc.Server) {
			meta.RegisterMetaServer(srv, g)
		}),
		server.WithGRPCOption(
			grpc.ChainUnaryInterceptor(grpc.RecoverInterceptor()),
			grpc.ChainStreamInterceptor(grpc.RecoverStreamInterceptor()),
		),
		server.WithPreStartFunc(func() error {
			// TODO check unbackupped upstream
			return nil
		}),
		server.WithPreStopFunction(func() error {
			// TODO backup all index data here
			return nil
		}),
	}

	var obs observability.Observability
	if cfg.Observability.Enabled {
		obs, err = observability.NewWithConfig(cfg.Observability)
		if err != nil {
			return nil, err
		}
		grpcServerOptions = append(
			grpcServerOptions,
			server.WithGRPCOption(
				grpc.StatsHandler(metric.NewServerHandler()),
			),
		)
	}

	srv, err := starter.New(
		starter.WithConfig(cfg.Server),
		starter.WithREST(func(sc *iconf.Server) []server.Option {
			return []server.Option{
				server.WithHTTPHandler(
					router.New(
						router.WithTimeout(sc.HTTP.HandlerTimeout),
						router.WithErrGroup(eg),
						router.WithHandler(
							rest.New(
								rest.WithMeta(g),
							),
						),
					)),
			}
		}),
		starter.WithGRPC(func(sc *iconf.Server) []server.Option {
			return grpcServerOptions
		}),
		// TODO add GraphQL handler
	)

	if err != nil {
		return nil, err
	}

	return &run{
		eg:            eg,
		cfg:           cfg,
		h:             h,
		server:        srv,
		observability: obs,
	}, nil
}

func (r *run) PreStart(ctx context.Context) error {
	err := r.h.Open(".halodb")
	if err != nil {
		return err
	}
	if r.observability != nil {
		return r.observability.PreStart(ctx)
	}
	return nil
}

func (r *run) Start(ctx context.Context) (<-chan error, error) {
	ech := make(chan error, 2)
	var oech, sech <-chan error
	r.eg.Go(safety.RecoverFunc(func() (err error) {
		defer close(ech)
		if r.observability != nil {
			oech = r.observability.Start(ctx)
		}
		sech = r.server.ListenAndServe(ctx)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err = <-oech:
			case err = <-sech:
			}
			if err != nil {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case ech <- err:
				}
			}
		}
	}))
	return ech, nil
}

func (r *run) PreStop(ctx context.Context) error {
	return nil
}

func (r *run) Stop(ctx context.Context) error {
	if r.observability != nil {
		r.observability.Stop(ctx)
	}
	return r.server.Shutdown(ctx)
}

func (r *run) PostStop(ctx context.Context) error {
	return r.h.Close()
}
