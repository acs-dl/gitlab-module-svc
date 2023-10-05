package processor

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) getLinkType(link string, priority int) (string, error) {
	checkType, err := gitlab.GetPermissionWithType(p.pqueues.SuperUserPQueue, any(p.gitlabClient.FindTypeFromApi), []any{any(link)}, priority)
	if err != nil {
		return "", err
	}

	if validation.Validate(checkType.Type, validation.In(data.Group, data.Project)) != nil {
		return "", errors.Wrap(err, "Failed to validate link type")
	}

	return checkType.Type, nil
}

func (p *processor) indexHasParentChild(gitlabId int64, link string, logger *logan.Entry) error {
	subPermission, err := p.subsQ.WithPermissions().FilterByGitlabIds(gitlabId).FilterByLinks(link).OrderBy("subs_link").Get()
	if err != nil {
		logger.WithError(err).Errorf("failed to get permission for `%d` from `%s`", gitlabId, link)
		return errors.Errorf("Failed to get permission for `%d` for `%s` from database", gitlabId, link)
	}

	if subPermission == nil {
		return errors.Errorf("No permission was found for `%d` for `%s` from database", gitlabId, link)
	}

	err = p.checkHasParent(*subPermission, logger)
	if err != nil {
		logger.WithError(err).Errorf("failed to check has parent")
		return err
	}

	return nil
}
