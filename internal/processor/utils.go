package processor

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) getLinkType(link string, priority int) (string, error) {
	checkType, err := gitlab.GetPermissionWithType(p.pqueues.SuperUserPQueue, any(p.gitlabClient.FindTypeFromApi), []any{any(link)}, priority)
	if err != nil {
		return "", errors.Wrap(err, "failed to get link type")
	}

	if checkType == nil {
		return "", errors.Errorf("no type was found for `%s`", link)
	}

	if validation.Validate(checkType.Type, validation.In(data.Group, data.Project)) != nil {
		return "", errors.Wrap(err, "something wrong with link type")
	}

	return checkType.Type, nil
}

func (p *processor) indexHasParentChild(gitlabId int64, link string) error {
	subPermission, err := p.subsQ.WithPermissions().FilterByGitlabIds(gitlabId).FilterByLinks(link).OrderBy("subs_link").Get()
	if err != nil {
		return errors.Wrap(err, "failed to get permission in permission db")
	}
	if subPermission == nil {
		return errors.Wrap(err, "got empty permission in permission db")
	}

	err = p.checkHasParent(*subPermission)
	if err != nil {
		return errors.Wrap(err, "failed to check parent level")
	}

	return nil
}
