package config

import (
	"context"
	"os"

	vault "github.com/hashicorp/vault/api"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GitLabCfg struct {
	SuperToken string `json:"super_token"`
	UsualToken string `json:"usual_token"`
}

func (c *config) Gitlab() *GitLabCfg {
	return c.gitlab.Do(func() interface{} {
		var cfg GitLabCfg

		client := createVaultClient()
		mountPath, secretPath := retrieveVaultPaths(c.getter)

		secret, err := client.KVv2(mountPath).Get(context.Background(), secretPath)
		if err != nil {
			panic(errors.Wrap(err, "failed to read from the vault"))
		}

		value, ok := secret.Data["super_token"].(string)
		if !ok {
			panic(errors.New("super token has wrong type"))
		}
		cfg.SuperToken = value

		value, ok = secret.Data["usual_token"].(string)
		if !ok {
			panic(errors.New("usual token has wrong type"))
		}
		cfg.UsualToken = value

		return &cfg
	}).(*GitLabCfg)
}

func createVaultClient() *vault.Client {
	vaultCfg := vault.DefaultConfig()
	vaultCfg.Address = os.Getenv("VAULT_ADDR")

	client, err := vault.NewClient(vaultCfg)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize a Vault client"))
	}

	client.SetToken(os.Getenv("VAULT_TOKEN"))

	return client
}

func retrieveVaultPaths(getter kv.Getter) (mount string, secret string) {
	type vCfg struct {
		MountPath  string `fig:"mount_path"`
		SecretPath string `fig:"secret_path"`
	}

	var vaultCfg vCfg

	err := figure.
		Out(&vaultCfg).
		With(figure.BaseHooks).
		From(kv.MustGetStringMap(getter, "vault")).
		Please()
	if err != nil {
		panic(errors.Wrap(err, "failed to figure out vault params from config"))
	}

	return vaultCfg.MountPath, vaultCfg.SecretPath
}
