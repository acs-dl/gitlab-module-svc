package gitlab

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (g *gitlab) FindTypeFromApi(link string) (*TypeSub, error) {
	group, err := g.getGroupFromApi(link)
	if err != nil {
		return nil, err
	}
	if group != nil {
		return &TypeSub{data.Group, *group}, err
	}

	project, err := g.getProjectFromApi(link)
	if err != nil {
		return nil, err
	}
	if project != nil {
		return &TypeSub{data.Project, *project}, nil
	}

	return nil, errors.Errorf("No project or group was found in Gitlab API by `%s` link", link)
}
