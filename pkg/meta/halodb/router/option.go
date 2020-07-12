package router

import (
	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/handler/rest"
	"github.com/rinx/vald-meta-halodb/internal/errgroup"
)

type Option func(*router)

var (
	defaultOpts = []Option{
		WithTimeout("3s"),
	}
)

func WithHandler(h rest.Handler) Option {
	return func(r *router) {
		r.handler = h
	}
}

func WithTimeout(timeout string) Option {
	return func(r *router) {
		r.timeout = timeout
	}
}

func WithErrGroup(eg errgroup.Group) Option {
	return func(r *router) {
		r.eg = eg
	}
}
