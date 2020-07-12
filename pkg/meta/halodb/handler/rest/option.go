package rest

import "github.com/vdaas/vald/apis/grpc/meta"

type Option func(*handler)

var (
	defaultOpts = []Option{}
)

func WithMeta(m meta.MetaServer) Option {
	return func(h *handler) {
		h.meta = m
	}
}
