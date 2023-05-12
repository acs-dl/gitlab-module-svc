package api

import (
	"context"
	"net/http"

	"github.com/acs-dl/gitlab-module-svc/internal/config"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Router struct {
	cfg config.Config
	ctx context.Context
}

func (r *Router) Run(_ context.Context) error {
	router := r.apiRouter()

	if err := r.cfg.Copus().RegisterChi(router); err != nil {
		return errors.Wrap(err, "cop failed")
	}

	return http.Serve(r.cfg.Listener(), router)
}

func NewRouterAsInterface(cfg config.Config, ctx context.Context) interface{} {
	return interface{}(&Router{
		cfg: cfg,
		ctx: ctx,
	})
}

func RunRouterAsInterface(structure interface{}, ctx context.Context) {
	err := (structure.(*Router)).Run(ctx)
	if err != nil {
		panic(err)
	}
}
