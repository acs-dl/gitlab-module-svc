package processor

import (
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) createHigherLevelPermissions(permission data.Sub) error {
	if permission.ParentId == nil { //must be handled before
		return errors.Errorf("parent id is empty")
	}

	sub, err := p.subsQ.FilterByIds(*permission.ParentId).Get()
	if err != nil {
		return errors.Wrap(err, "failed to get parent sub")
	}

	if sub == nil {
		return errors.Errorf("sub is empty")
	}

	err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{ParentLink: &sub.Link})
	if err != nil {
		return errors.Wrap(err, "failed to update parent link")
	}

	for sub != nil {
		newPermission := *permission.Permission
		newPermission.Link = sub.Link
		newPermission.Type = sub.Type
		newPermission.AccessLevel = 0
		newPermission.CreatedAt = time.Now()
		newPermission.ExpiresAt = time.Time{}

		if sub.ParentId == nil {
			//we reached the highest level
			newPermission.ParentLink = nil
			err = p.permissionsQ.Upsert(newPermission)
			if err != nil {
				return errors.Wrap(err, "failed to upsert permission")
			}

			hasParent := false
			hasChild := true
			err = p.permissionsQ.
				FilterByGitlabIds(newPermission.GitlabId).
				FilterByLinks(newPermission.Link).
				Update(data.PermissionToUpdate{
					HasParent: &hasParent,
					HasChild:  &hasChild,
				})
			if err != nil {
				return errors.Wrap(err, "failed to update has child and has parent")
			}
			break
		}

		sub, err = p.subsQ.FilterByIds(*sub.ParentId).Get()
		if err != nil {
			return errors.Wrap(err, "failed to get sub")
		}

		if sub == nil {
			return errors.Errorf("sub is empty")
		}

		newPermission.ParentLink = &sub.Link
		err = p.permissionsQ.Upsert(newPermission)
		if err != nil {
			return errors.Wrap(err, "failed to upsert permission")
		}

		hasParent := false
		hasChild := true
		err = p.permissionsQ.
			FilterByGitlabIds(newPermission.GitlabId).
			FilterByLinks(newPermission.Link).
			Update(data.PermissionToUpdate{
				HasParent: &hasParent,
				HasChild:  &hasChild,
			})
		if err != nil {
			return errors.Wrap(err, "failed to update has child and has parent")
		}
	}

	return nil
}
