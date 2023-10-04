package processor

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) checkHasParent(permission data.Sub, logger *logan.Entry) error {
	if permission.ParentId == nil {
		hasParent := false
		err := p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
		if err != nil {
			logger.WithError(err).Errorf("failed to update has parent for `%d` in `%s`", permission.GitlabId, permission.Link)
			return errors.Errorf("Failed to update has parent field for `%d` in `%s` in database", permission.GitlabId, permission.Link)
		}

		return nil
	}

	parentPermission, err := p.subsQ.WithPermissions().FilterByGitlabIds(permission.GitlabId).FilterByIds(*permission.ParentId).OrderBy("subs_link").Get()
	if err != nil {
		logger.WithError(err).Errorf("failed to get permission for `%d` in `%d`", permission.GitlabId, *permission.ParentId)
		return errors.Errorf("Failed to get permission for `%d` in `%d` in database", permission.GitlabId, *permission.ParentId)
	}

	if parentPermission == nil || parentPermission.AccessLevel == 0 {
		//suppose that it means: that user is not in parent repo only in lower level
		err = p.createHigherLevelPermissions(permission, logger)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create higher permissions")
			return err
		}

		hasParent := false
		err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
		if err != nil {
			logger.WithError(err).Errorf("failed to update has parent for `%d` in `%s`", permission.GitlabId, permission.Link)
			return errors.Errorf("Failed to update has parent field for `%d` in `%s` in database", permission.GitlabId, permission.Link)
		}

		return nil
	}

	err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{ParentLink: &parentPermission.Link})
	if err != nil {
		logger.WithError(err).Errorf("failed to update parent link for `%d` in `%s`", permission.GitlabId, permission.Link)
		return errors.Errorf("Failed to update parent link field for `%d` in `%s` in database", permission.GitlabId, permission.Link)
	}

	if permission.AccessLevel == parentPermission.AccessLevel {
		return nil
	}

	hasParent := false
	err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
	if err != nil {
		logger.WithError(err).Errorf("failed to update has parent for `%d` in `%s`", permission.GitlabId, permission.Link)
		return errors.Errorf("Failed to update has parent field for `%d` in `%s` in database", permission.GitlabId, permission.Link)
	}

	hasChild := true
	err = p.permissionsQ.FilterByGitlabIds(parentPermission.GitlabId).FilterByLinks(parentPermission.Link).Update(data.PermissionToUpdate{HasChild: &hasChild})
	if err != nil {
		logger.WithError(err).Errorf("failed to update has child for `%d` in `%s`", parentPermission.GitlabId, parentPermission.Link)
		return errors.Errorf("Failed to update has child field for `%d` in `%s` in database", parentPermission.GitlabId, parentPermission.Link)
	}

	return nil
}
