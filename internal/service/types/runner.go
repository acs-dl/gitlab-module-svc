package types

import (
	"context"

	"github.com/acs-dl/gitlab-module-svc/internal/config"
)

type Runner = func(context context.Context, config config.Config)
