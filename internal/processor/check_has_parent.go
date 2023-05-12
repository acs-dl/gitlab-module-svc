package processor

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) checkHasParent(permission data.Sub) error {
	if permission.ParentId == nil {
		hasParent := false
		err := p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
		if err != nil {
			return errors.Wrap(err, "failed to has parent")
		}

		return nil
	}

	parentPermission, err := p.subsQ.WithPermissions().FilterByGitlabIds(permission.GitlabId).FilterByIds(*permission.ParentId).OrderBy("subs_link").Get()
	if err != nil {
		return errors.Wrap(err, "failed to get parent permission")
	}

	if parentPermission == nil || parentPermission.AccessLevel == 0 {
		//suppose that it means: that user is not in parent repo only in lower level
		err = p.createHigherLevelPermissions(permission)
		if err != nil {
			return errors.Wrap(err, "failed to create higher parent permissions")
		}

		hasParent := false
		err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
		if err != nil {
			return errors.Wrap(err, "failed to update has parent")
		}

		return nil
	}

	err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{ParentLink: &parentPermission.Link})
	if err != nil {
		return errors.Wrap(err, "failed to update parent link")
	}

	if permission.AccessLevel == parentPermission.AccessLevel {
		return nil
	}

	hasParent := false
	err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{HasParent: &hasParent})
	if err != nil {
		return errors.Wrap(err, "failed to update has parent")
	}

	hasChild := true
	err = p.permissionsQ.FilterByGitlabIds(parentPermission.GitlabId).FilterByLinks(parentPermission.Link).Update(data.PermissionToUpdate{HasChild: &hasChild})
	if err != nil {
		return errors.Wrap(err, "failed to update has child")
	}

	return nil
}
