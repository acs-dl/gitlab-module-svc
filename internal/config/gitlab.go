package config

import (
	knox "gitlab.com/distributed_lab/knox/knox-fork/client/external_kms"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GitLabCfg struct {
	SuperToken string `json:"super_token"`
	UsualToken string `json:"usual_token"`
}

func (c *config) Gitlab() *GitLabCfg {
	return c.gitlab.Do(func() interface{} {
		var cfg GitLabCfg

		client := knox.NewKeyManagementClient(c.getter)

		key, err := client.GetKey("super_token", "4697000974761249000")
		if err != nil {
			panic(errors.Wrap(err, "failed to get super token key"))
		}

		cfg.SuperToken = string(key[:])

		key, err = client.GetKey("usual_token", "1959184036590655200")
		if err != nil {
			panic(errors.Wrap(err, "failed to get usual token key"))
		}
		cfg.UsualToken = string(key[:])

		return &cfg
	}).(*GitLabCfg)
}
