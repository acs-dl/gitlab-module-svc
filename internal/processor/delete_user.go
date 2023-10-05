package processor

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateDeleteUser(msg data.ModulePayload) error {
	return validation.Errors{
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

// HandleDeleteUserAction .Wrap from return err was removed because
// errors must be human-readable from very low level to send them in FE.
// log must be put before every error to track it if any
func (p *processor) HandleDeleteUserAction(msg data.ModulePayload) error {
	log := p.log.WithField("message", msg.RequestId)
	log.Infof("start handling delete user action")

	err := p.validateDeleteUser(msg)
	if err != nil {
		log.WithError(err).Errorf("failed to validate fields")
		return errors.Wrap(err, "Request is not valid")
	}

	userApi, err := gitlab.GetUser(p.pqueues.UserPQueue, any(p.gitlabClient.GetUserFromApi), []any{any(msg.Username)}, pqueue.NormalPriority)
	if err != nil {
		log.WithError(err).Errorf("failed to get user from API")
		return err
	}

	permissions, err := p.permissionsQ.
		FilterByGitlabIds(userApi.GitlabId).
		Select()
	if err != nil {
		log.WithError(err).Errorf("failed to select permissions by gitlab id `%d`", userApi.GitlabId)
		return errors.Errorf("Failed to select permissions by `%s` Gitlab ID", userApi.GitlabId)
	}

	for _, permission := range permissions {
		err = p.removePermissionFromRemoteAndLocal(permission, log)
		if err != nil {
			log.WithError(err).Errorf("failed to remove permission from remote and local")
			return err
		}
	}

	err = p.removeUserFromService(msg.RequestId, userApi.GitlabId, log)
	if err != nil {
		log.WithError(err).Errorf("failed to delete user from service")
		return err
	}

	log.Infof("finish handling delete user message action")
	return nil
}

func (p *processor) removePermissionFromRemoteAndLocal(permission data.Permission, logger *logan.Entry) error {
	permissionApi, err := gitlab.GetPermission(p.pqueues.UserPQueue, any(p.gitlabClient.CheckUserFromApi), []any{
		any(permission.Link),
		any(permission.Type),
		any(permission.GitlabId),
	}, pqueue.NormalPriority)
	if err != nil {
		logger.WithError(err).Errorf("failed to check `%d` from API in `%s`", permission.GitlabId, permission.Link)
		return err
	}

	if permissionApi != nil && !permission.HasParent && permission.AccessLevel != 0 {
		err = gitlab.GetRequestError(p.pqueues.SuperUserPQueue, any(p.gitlabClient.RemoveUserFromApi), []any{
			any(permission.Link),
			any(permission.Type),
			any(permission.GitlabId),
		}, pqueue.NormalPriority)
		if err != nil {
			logger.WithError(err).Errorf("failed to remove `%d` from `%s` via gitlab API", permission.GitlabId, permission.Link)
			return err
		}
	}

	err = p.permissionsQ.FilterByLinks(permission.Link).FilterByTypes(permission.Type).FilterByGitlabIds(permission.GitlabId).Delete()
	if err != nil {
		logger.WithError(err).Errorf("failed to delete permission for `%d` from `%s` from db", permission.GitlabId, permission.Link)
		return errors.Errorf("Failed to delete permission for `%d` from `%s` from database", permission.GitlabId, permission.Link)
	}

	return nil
}

func (p *processor) removeUserFromService(requestId string, gitlabId int64, logger *logan.Entry) error {
	dbUser, err := p.usersQ.FilterByGitlabIds(gitlabId).Get()
	if err != nil {
		logger.WithError(err).Errorf("failed to get user by `%s` gitlab id", gitlabId)
		return errors.Errorf("Failed to get user `%s` by Gitlab ID from database", gitlabId)
	}

	if dbUser == nil {
		return errors.Errorf("No user with Gitlab ID `%d` was found in database", gitlabId)
	}

	err = p.usersQ.FilterByGitlabIds(gitlabId).Delete()
	if err != nil {
		logger.WithError(err).Errorf("failed to delete user by `%s` gitlab id", gitlabId)
		return errors.Errorf("Failed to delete user `%s` by Gitlab ID from database", gitlabId)
	}

	if dbUser.Id == nil {
		err = p.SendDeleteUser(requestId, *dbUser)
		if err != nil {
			logger.WithError(err).Errorf("failed to publish delete user")
			return err
		}
	}

	return nil
}
