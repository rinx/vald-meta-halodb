package grpc

import "github.com/rinx/vald-meta-halodb/pkg/meta/halodb/service"

type Option func(*server)

var (
	defaultOpts = []Option{}
)

func WithHaloDB(h service.HaloDB) Option {
	return func(s *server) {
		s.haloDB = h
	}
}
