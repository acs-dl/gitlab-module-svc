package processor

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateDeleteUser(msg data.ModulePayload) error {
	return validation.Errors{
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

func (p *processor) HandleDeleteUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateDeleteUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	userApi, err := gitlab.GetUser(p.pqueues.UserPQueue, any(p.gitlabClient.GetUserFromApi), []any{any(msg.Username)}, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user id from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting user id from api")
	}
	if userApi == nil {
		p.log.Errorf("no user was found from api for message action with id `%s`", msg.RequestId)
		return errors.New("no user was found from api")
	}

	permissions, err := p.permissionsQ.
		FilterByGitlabIds(userApi.GitlabId).
		FilterByHasParent(false).
		Select()
	if err != nil {
		p.log.WithError(err).Errorf("failed to get permissions by gitlab id `%d` for message action with id `%s`", userApi.GitlabId, msg.RequestId)
		return errors.Wrap(err, "failed to delete permission")
	}

	for _, permission := range permissions {
		err = p.removePermissionFromRemoteAndLocal(permission)
		if err != nil {
			p.log.WithError(err).Errorf("failed to remove permission from remote and local for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to remove permission from remote and local")
		}
	}

	err = p.removeUserFromService(msg.RequestId, userApi.GitlabId)
	if err != nil {
		p.log.WithError(err).Errorf("failed to make remove user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make remove user transaction")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}

func (p *processor) removePermissionFromRemoteAndLocal(permission data.Permission) error {
	permissionApi, err := gitlab.GetPermission(p.pqueues.UserPQueue, any(p.gitlabClient.CheckUserFromApi), []any{
		any(permission.Link),
		any(permission.Type),
		any(permission.GitlabId),
	}, pqueue.NormalPriority)
	if err != nil {
		return errors.Wrap(err, "some error while checking user from api")
	}

	if permissionApi != nil {
		err = gitlab.GetRequestError(p.pqueues.UserPQueue, any(p.gitlabClient.RemoveUserFromApi), []any{
			any(permission.Link),
			any(permission.Type),
			any(permission.GitlabId),
		}, pqueue.NormalPriority)
		if err != nil {
			return errors.Wrap(err, "some error while removing user from api")
		}
	}

	err = p.permissionsQ.FilterByLinks(permission.Link).FilterByTypes(permission.Type).FilterByGitlabIds(permission.GitlabId).Delete()
	if err != nil {
		return errors.Wrap(err, "failed to delete permission")
	}

	return nil
}

func (p *processor) removeUserFromService(requestId string, gitlabId int64) error {
	dbUser, err := p.usersQ.FilterByGitlabIds(gitlabId).Get()
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	if dbUser == nil {
		return errors.New("something wrong with db user")
	}

	err = p.usersQ.FilterByGitlabIds(gitlabId).Delete()
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if dbUser.Id == nil {
		err = p.SendDeleteUser(requestId, *dbUser)
		if err != nil {
			return errors.Wrap(err, "failed to publish delete user")
		}
	}

	return nil
}
