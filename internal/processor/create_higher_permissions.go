package processor

import (
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) createHigherLevelPermissions(permission data.Sub, logger *logan.Entry) error {
	if permission.ParentId == nil { //must be handled before
		return errors.Errorf("Failed to create higher level permissions: Parent id is empty")
	}

	sub, err := p.subsQ.FilterByIds(*permission.ParentId).Get()
	if err != nil {
		logger.WithError(err).Errorf("failed to get sub by `%d` id", *permission.ParentId)
		return errors.Errorf("Failed to get link details by `%d` id from database", *permission.ParentId)
	}

	if sub == nil {
		logger.Errorf("sub for `%d` is empty", *permission.ParentId)
		return errors.Errorf("Link details for `%d` is empty", *permission.ParentId)
	}

	err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(permission.Link).Update(data.PermissionToUpdate{ParentLink: &sub.Link})
	if err != nil {
		logger.WithError(err).Errorf("failed to update parent link for `%d` in `%s`", permission.GitlabId, permission.Link)
		return errors.Errorf("Failed to update parent link field for `%d` in `%s` in database", permission.GitlabId, permission.Link)
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
			err = upsertHigherLevelPermission(newPermission, p.permissionsQ, logger)
			if err != nil {
				return err
			}

			break
		}

		subID := *sub.ParentId
		sub, err = p.subsQ.FilterByIds(subID).Get()
		if err != nil {
			logger.WithError(err).Errorf("failed to get sub by `%d` id", subID)
			return errors.Errorf("Failed to get link details by `%d` id from database", subID)
		}

		if sub == nil {
			logger.Errorf("sub for `%d` is empty", subID)
			return errors.Errorf("Link details for `%d` is empty", subID)
		}

		newPermission.ParentLink = &sub.Link
		err = upsertHigherLevelPermission(newPermission, p.permissionsQ, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func upsertHigherLevelPermission(permission data.Permission, permissionsQ data.Permissions, logger *logan.Entry) error {
	err := permissionsQ.Upsert(permission)
	if err != nil {
		logger.WithError(err).Errorf("failed to upsert  new permission")
		return errors.Errorf("Failed to create new higher level permission")
	}

	hasParent := false
	hasChild := true
	err = permissionsQ.
		FilterByGitlabIds(permission.GitlabId).
		FilterByLinks(permission.Link).
		Update(data.PermissionToUpdate{
			HasParent: &hasParent,
			HasChild:  &hasChild,
		})
	if err != nil {
		logger.WithError(err).Errorf("failed to update has parent and has child for `%d` in `%s`", permission.GitlabId, permission.Link)
		return errors.Errorf("Failed to update has parent and has child fields for `%d` in `%s` in database", permission.GitlabId, permission.Link)
	}

	return nil
}
