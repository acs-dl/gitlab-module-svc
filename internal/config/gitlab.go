package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GitLabCfg struct {
	SuperToken string `figure:"super_token"`
	UsualToken string `figure:"usual_token"`
}

func (c *config) Gitlab() *GitLabCfg {
	return c.gitlab.Do(func() interface{} {
		var cfg GitLabCfg
		err := figure.
			Out(&cfg).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, "gitlab")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out gitlab params from config"))
		}

		return &cfg
	}).(*GitLabCfg)
}
