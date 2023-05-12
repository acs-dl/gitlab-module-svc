package gitlab

import (
	"context"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/acs-dl/gitlab-module-svc/internal/config"
	"github.com/acs-dl/gitlab-module-svc/internal/data"
)

type GitlabClient interface {
	GetUserFromApi(username string) (*data.User, error)

	FindTypeFromApi(link string) (*TypeSub, error)

	AddUsersFromApi(link, typeTo string, info data.Permission) (*data.Permission, error)
	GetUsersFromApi(link, typeTo string) ([]data.Permission, error)
	RemoveUserFromApi(link, typeTo string, gitlabId int64) error
	UpdateUserFromApi(info data.Permission) (*data.Permission, error)

	CheckUserFromApi(link, typeTo string, userId int64) (*data.Permission, error)
	SearchByFromApi(username string) ([]data.User, error)

	GetSubgroupsFomApi(link string) ([]data.Sub, error)
	GetProjectsFomApi(link string) ([]data.Sub, error)
}

type TypeSub struct {
	Type string
	Sub  data.Sub
}

type gitlab struct {
	superUserToken string
	userToken      string
	log            *logan.Entry
}

func NewGitlabAsInterface(cfg config.Config, _ context.Context) interface{} {
	return interface{}(&gitlab{
		superUserToken: cfg.Gitlab().SuperToken,
		userToken:      cfg.Gitlab().UsualToken,
		log:            cfg.Log(),
	})
}

func GitlabClientInstance(ctx context.Context) GitlabClient {
	return ctx.Value("gitlab").(GitlabClient)
}

func CtxGitlabClientInstance(entry interface{}, ctx context.Context) context.Context {
	return context.WithValue(ctx, "gitlab", entry)
}
