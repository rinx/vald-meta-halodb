package router

import (
	"net/http"

	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/handler/rest"
	"github.com/rinx/vald-meta-halodb/internal/errgroup"
	"github.com/rinx/vald-meta-halodb/internal/net/http/middleware"
	"github.com/rinx/vald-meta-halodb/internal/net/http/routing"
)

type router struct {
	handler rest.Handler
	eg      errgroup.Group
	timeout string
}

// New returns REST route&method information from handler interface
func New(opts ...Option) http.Handler {
	r := new(router)

	for _, opt := range append(defaultOpts, opts...) {
		opt(r)
	}

	h := r.handler

	return routing.New(
		routing.WithMiddleware(
			middleware.NewTimeout(
				middleware.WithTimeout(r.timeout),
				middleware.WithErrorGroup(r.eg),
			)),
		routing.WithRoutes([]routing.Route{
			{
				"Index",
				[]string{
					http.MethodGet,
				},
				"/",
				h.Index,
			},
			{
				"GetMeta",
				[]string{
					http.MethodGet,
				},
				"/meta",
				h.GetMeta,
			},
			{
				"GetMetas",
				[]string{
					http.MethodGet,
				},
				"/metas",
				h.GetMetas,
			},
			{
				"GetMetaInverse",
				[]string{
					http.MethodGet,
				},
				"/inverse/meta",
				h.GetMetaInverse,
			},
			{
				"GetMetasInverse",
				[]string{
					http.MethodGet,
				},
				"/inverse/metas",
				h.GetMetasInverse,
			},
			{
				"SetMeta",
				[]string{
					http.MethodPost,
				},
				"/meta",
				h.SetMeta,
			},

			{
				"SetMetas",
				[]string{
					http.MethodPost,
				},
				"/metas",
				h.SetMetas,
			},
			{
				"DeleteMeta",
				[]string{
					http.MethodPost,
				},
				"/meta",
				h.DeleteMeta,
			},
			{
				"DeleteMetas",
				[]string{
					http.MethodPost,
				},
				"/metas",
				h.DeleteMetas,
			},
			{
				"DeleteMetaInverse",
				[]string{
					http.MethodPost,
				},
				"/inverse/meta",
				h.DeleteMetaInverse,
			},
			{
				"DeleteMetasInverse",
				[]string{
					http.MethodPost,
				},
				"/inverse/metas",
				h.DeleteMetasInverse,
			},
		}...))
}
