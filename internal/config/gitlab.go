package config

import (
	"os"

	validation "github.com/go-ozzo/ozzo-validation"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GitLabCfg struct {
	SuperToken string `json:"super_token"`
	UsualToken string `json:"usual_token"`
}

func (c *config) Gitlab() *GitLabCfg {
	return c.gitlab.Do(func() interface{} {
		cfg := lookupConfigEnv()

		err := cfg.validate()
		if err != nil {
			panic(errors.Wrap(err, "failed to get usual token key"))
		}

		return cfg
	}).(*GitLabCfg)
}

func lookupConfigEnv() *GitLabCfg {
	superToken, ok := os.LookupEnv("super_token")
	if !ok {
		panic(errors.New("no super_token env variable"))
	}

	usualToken, ok := os.LookupEnv("usual_token")
	if !ok {
		panic(errors.New("no usual_token env variable"))
	}

	return &GitLabCfg{
		superToken,
		usualToken,
	}
}

func (g *GitLabCfg) validate() error {
	return validation.Errors{
		"super_token": validation.Validate(g.SuperToken, validation.Required),
		"user_token":  validation.Validate(g.UsualToken, validation.Required),
	}.Filter()
}
