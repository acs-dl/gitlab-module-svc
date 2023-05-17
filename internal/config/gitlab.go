package config

import (
	"encoding/json"
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
		var cfg GitLabCfg
		value, ok := os.LookupEnv("gitlab")
		if !ok {
			panic(errors.New("no gitlab env variable"))
		}
		err := json.Unmarshal([]byte(value), &cfg)
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out gitlab params from env variable"))
		}

		err = cfg.validate()
		if err != nil {
			panic(errors.Wrap(err, "failed to validate gitlab params"))
		}
		return &cfg
	}).(*GitLabCfg)
}

func (g *GitLabCfg) validate() error {
	return validation.Errors{
		"super_token": validation.Validate(g.SuperToken, validation.Required),
		"user_token":  validation.Validate(g.UsualToken, validation.Required),
	}.Filter()
}
