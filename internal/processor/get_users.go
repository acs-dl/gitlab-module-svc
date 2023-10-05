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

// HandleGetUsersAction .Wrap from return err was removed because
// errors must be human-readable from very low level to send them in FE.
// log must be put before every error to track it if any
func (p *processor) HandleGetUsersAction(msg data.ModulePayload) error {
	log := p.log.WithField("message", msg.RequestId)
	log.Infof("start handling get users action")

	err := p.validateGetUsers(msg)
	if err != nil {
		log.WithError(err).Errorf("failed to validate fields")
		return errors.Wrap(err, "Request is not valid")
	}

	msg.Type, err = p.getLinkType(msg.Link, pqueue.LowPriority)
	if err != nil {
		log.WithError(err).Errorf("failed to get link type")
		return err
	}

	permissions, err := gitlab.GetPermissions(p.pqueues.SuperUserPQueue, any(p.gitlabClient.GetUsersFromApi), []any{
		any(msg.Link),
		any(msg.Type),
	}, pqueue.LowPriority)
	if err != nil {
		log.WithError(err).Errorf("failed to get user from API")
		return err
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
				log.WithError(err).Errorf("failed to create user in user db")
				return errors.Errorf("Failed to create new user in database")
			}

			usrDb, err := p.usersQ.FilterByUsernames(permission.Username).Get()
			if err != nil {
				log.WithError(err).Errorf("failed to get user `%s` from db", permission.Username)
				return errors.Errorf("Failed to get user `%s` from database", permission.Username)
			}

			if usrDb == nil {
				log.WithError(err).Errorf("no user `%s` in db", permission.Username)
				return errors.Errorf("No user `%s` in database was found", permission.Username)
			}

			usersToUnverified = append(usersToUnverified, *usrDb)

			permission.UserId = usrDb.Id
			permission.Type = msg.Type
			permission.Link = msg.Link
			permission.RequestId = msg.RequestId

			permission.ExpiresAt, err = helpers.ParseTime(permission.ExpiresString)
			if err != nil {
				log.WithError(err).Errorf("failed to parse time `%s` from API ", permission.ExpiresString)
				return err
			}

			err = p.permissionsQ.Upsert(permission)
			if err != nil {
				log.WithError(err).Errorf("failed to create permission in permission db")
				return errors.Errorf("Failed to create new user permission in database")
			}

			err = p.indexHasParentChild(permission.GitlabId, permission.Link, log)
			if err != nil {
				log.WithError(err).Errorf("failed to index has parent/child for `%d` from `%s`", permission.GitlabId, permission.Link)
				return err
			}

			return nil
		})
		if err != nil {
			log.WithError(err).Errorf("failed to make get users transaction")
			return err
		}
	}

	err = p.sendUsers(msg.RequestId, usersToUnverified)
	if err != nil {
		log.WithError(err).Errorf("failed to publish users to unverified-svc")
		return errors.Errorf("Failed to publish users to `unverified-svc`")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}
