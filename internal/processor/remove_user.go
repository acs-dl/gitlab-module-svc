package processor

import (
	"strings"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateRemoveUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":     validation.Validate(msg.Link, validation.Required),
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

func (p *processor) HandleRemoveUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateRemoveUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}
	msg.Link = strings.ToLower(msg.Link)

	msg.Type, err = p.getLinkType(msg.Link, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get link type for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get link type")
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

	dbUser, err := p.usersQ.FilterByUsernames(msg.Username).Get()
	if err != nil {
		p.log.WithError(err).Errorf("failed to get user from user db for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get user from user db")
	}

	if dbUser == nil {
		p.log.Errorf("no user with such username for message action with id `%s`", msg.RequestId)
		return errors.New("no user with such username")
	}

	err = gitlab.GetRequestError(p.pqueues.UserPQueue, any(p.gitlabClient.RemoveUserFromApi), []any{
		any(msg.Link),
		any(msg.Type),
		any(userApi.GitlabId),
	}, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to remove user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while removing user from api")
	}

	err = p.managerQ.Transaction(func() error {
		err = p.deleteLowerLevelPermissions(userApi.GitlabId, msg.Link, msg.Type)
		if err != nil {
			p.log.WithError(err).Errorf("failed to delete permission from db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to delete permission")
		}

		permissions, err := p.permissionsQ.FilterByGitlabIds(userApi.GitlabId).Select()
		if err != nil {
			p.log.WithError(err).Errorf("failed to get permissions by gitlab id `%d` for message action with id `%s`", userApi.GitlabId, msg.RequestId)
			return errors.Wrap(err, "failed to delete permission")
		}
		if len(permissions) == 0 {
			err = p.usersQ.FilterByGitlabIds(userApi.GitlabId).Delete()
			if err != nil {
				p.log.WithError(err).Errorf("failed to delete user by gitlab id `%d` for message action with id `%s`", userApi.GitlabId, msg.RequestId)
				return errors.Wrap(err, "failed to delete user")
			}

			if dbUser.Id == nil {
				err = p.SendDeleteUser(msg.RequestId, *dbUser)
				if err != nil {
					p.log.WithError(err).Errorf("failed to publish delete user for message action with id `%s`", msg.RequestId)
					return errors.Wrap(err, "failed to publish delete user")
				}
			}
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make remove user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make remove user transaction")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}

func (p *processor) deleteLowerLevelPermissions(gitlabId int64, link, typeTo string) error {
	err := p.permissionsQ.FilterByGitlabIds(gitlabId).FilterByTypes(typeTo).FilterByLinks(link).Delete()
	if err != nil {
		return errors.Wrap(err, "failed to delete permission")
	}

	permissions, err := p.permissionsQ.FilterByParentLinks(link).FilterByGitlabIds(gitlabId).Select()
	if err != nil {
		return errors.Wrap(err, "failed to select permissions")
	}

	if len(permissions) == 0 {
		return nil
	}

	for _, permission := range permissions {
		err = p.deleteLowerLevelPermissions(permission.GitlabId, permission.Link, permission.Type)
		if err != nil {
			return errors.Wrap(err, "failed to delete lower level permission")
		}
	}

	return nil
}
