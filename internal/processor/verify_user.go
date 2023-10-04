package processor

import (
	"strconv"
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateVerifyUser(msg data.ModulePayload) error {
	return validation.Errors{
		"user_id":  validation.Validate(msg.UserId, validation.Required),
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

func (p *processor) HandleVerifyUserAction(msg data.ModulePayload) error {
	log := p.log.WithField("message", msg.RequestId)
	log.Infof("start handling verify user action")

	err := p.validateVerifyUser(msg)
	if err != nil {
		log.WithError(err).Errorf("failed to validate fields")
		return errors.Wrap(err, "Request is not valid")
	}

	userId, err := strconv.ParseInt(msg.UserId, 10, 64)
	if err != nil {
		log.WithError(err).Errorf("failed to parse user id `%s`", msg.UserId)
		return errors.Errorf("Failed to parse user ID `%s`", msg.UserId)
	}

	userApi, err := gitlab.GetUser(p.pqueues.UserPQueue, any(p.gitlabClient.GetUserFromApi), []any{any(msg.Username)}, pqueue.NormalPriority)
	if err != nil {
		log.WithError(err).Errorf("failed to get user from API")
		return err
	}

	userApi.Id = &userId
	userApi.CreatedAt = time.Now()

	err = p.managerQ.Transaction(func() error {
		if err = p.usersQ.Upsert(*userApi); err != nil {
			log.WithError(err).Errorf("failed to create user in user db")
			return errors.Errorf("Failed to create new user in database")
		}

		err = p.permissionsQ.FilterByGitlabIds(userApi.GitlabId).Update(data.PermissionToUpdate{UserId: &userId})
		if err != nil {
			log.WithError(err).Errorf("failed to update user id for `%d` in permission db", msg.GitlabId)
			return errors.Errorf("Failed to update user id in user database for `%d` Gitlab ID", msg.GitlabId)
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Errorf("failed to make verify user transaction")
		return err
	}

	err = p.SendDeleteUser(msg.RequestId, *userApi)
	if err != nil {
		log.WithError(err).Errorf("failed to publish delete users")
		return err
	}

	log.Infof("finish handling add user message action")
	return nil
}
