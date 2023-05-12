package processor

import (
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/helpers"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateGetUsers(msg data.ModulePayload) error {
	return validation.Errors{
		"link": validation.Validate(msg.Link, validation.Required),
	}.Filter()
}

func (p *processor) HandleGetUsersAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateGetUsers(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	msg.Type, err = p.getLinkType(msg.Link, pqueue.LowPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get link type for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to get link type")
	}

	permissions, err := gitlab.GetPermissions(p.pqueues.SuperUserPQueue, any(p.gitlabClient.GetUsersFromApi), []any{
		any(msg.Link),
		any(msg.Type),
	}, pqueue.LowPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get users from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while getting users from api")
	}

	usersToUnverified := make([]data.User, 0)

	for _, permission := range permissions {
		err = p.managerQ.Transaction(func() error {
			if err = p.usersQ.Upsert(data.User{
				GitlabUsername: permission.Username,
				GitlabId:       permission.GitlabId,
				CreatedAt:      time.Now(),
				AvatarUrl:      permission.AvatarUrl,
				Name:           permission.Name,
			}); err != nil {
				p.log.WithError(err).Errorf("failed to creat user in user db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to create user in user db")
			}

			usrDb, err := p.usersQ.FilterByUsernames(permission.Username).Get()
			if err != nil {
				p.log.WithError(err).Errorf("failed to get user form user db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to get user from user db")
			}

			if usrDb == nil {
				p.log.WithError(err).Errorf("no such user in db for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "no such user in db")
			}

			usersToUnverified = append(usersToUnverified, *usrDb)

			permission.UserId = usrDb.Id
			permission.Type = msg.Type
			permission.Link = msg.Link
			permission.RequestId = msg.RequestId

			permission.ExpiresAt, err = helpers.ParseTime(permission.ExpiresString)
			if err != nil {
				p.log.WithError(err).Errorf("failed to parse time from API for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "some error while parsing time from api")
			}

			err = p.permissionsQ.Upsert(permission)
			if err != nil {
				p.log.WithError(err).Errorf("failed to upsert permission for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to upsert permission in permission db")
			}

			err = p.indexHasParentChild(permission.GitlabId, permission.Link)
			if err != nil {
				p.log.WithError(err).Errorf("failed to check has parent/child for message action with id `%s`", msg.RequestId)
				return errors.Wrap(err, "failed to check parent level")
			}

			return nil
		})
		if err != nil {
			p.log.WithError(err).Errorf("failed to make get users transaction for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to make get users transaction")
		}
	}

	err = p.sendUsers(msg.RequestId, usersToUnverified)
	if err != nil {
		p.log.WithError(err).Errorf("failed to publish users for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to publish users")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
