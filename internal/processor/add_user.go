package processor

import (
	"strconv"
	"strings"
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/helpers"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateAddUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":         validation.Validate(msg.Link, validation.Required),
		"username":     validation.Validate(msg.Username, validation.Required),
		"user_id":      validation.Validate(msg.UserId, validation.Required),
		"access_level": validation.Validate(msg.AccessLevel, validation.Required),
	}.Filter()
}

func (p *processor) HandleAddUserAction(msg data.ModulePayload) error {
	log := p.log.WithField("message", msg.RequestId)
	log.Infof("start handling add user action")

	err := p.validateAddUser(msg)
	if err != nil {
		log.WithError(err).Errorf("failed to validate fields")
		return errors.Wrap(err, "Request is not valid")
	}

	msg.Link = strings.ToLower(msg.Link)
	userId, err := strconv.ParseInt(msg.UserId, 10, 64)
	if err != nil {
		log.WithError(err).Errorf("failed to parse user id `%s`", msg.UserId)
		return errors.Errorf("Failed to parse user ID `%s`", msg.UserId)
	}

	permission, err := p.addUser(msg.Username, msg.Link, msg.AccessLevel, log)
	if err != nil {
		log.WithError(err).Errorf("failed to add user from API")
		return err
	}

	permission.UserId = &userId
	permission.Link = msg.Link
	permission.RequestId = msg.RequestId
	permission.CreatedAt = time.Now()

	permission.ExpiresAt, err = helpers.ParseTime(permission.ExpiresString)
	if err != nil {
		log.WithError(err).Errorf("failed to parse time `%s` from API ", permission.ExpiresString)
		return err
	}

	dbUser := data.User{
		Id:             &userId,
		GitlabUsername: permission.Username,
		GitlabId:       permission.GitlabId,
		CreatedAt:      permission.CreatedAt,
		Name:           permission.Name,
		AvatarUrl:      permission.AvatarUrl,
	}

	err = p.managerQ.Transaction(func() error {
		if err = p.usersQ.Upsert(dbUser); err != nil {
			log.WithError(err).Errorf("failed to create user in user db")
			return errors.Errorf("Failed to create new user in database")
		}

		if err = p.permissionsQ.Upsert(*permission); err != nil {
			log.WithError(err).Errorf("failed to create permission in permission db")
			return errors.Errorf("Failed to create new user permission in database")
		}

		//in case if we have some rows without id from identity
		err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).Update(data.PermissionToUpdate{UserId: permission.UserId})
		if err != nil {
			log.WithError(err).Errorf("failed to update user id for `%d` in permission db", msg.GitlabId)
			return errors.Errorf("Failed to update user id in user database for `%d` Gitlab ID", msg.GitlabId)
		}

		err = p.indexHasParentChild(dbUser.GitlabId, msg.Link, log)
		if err != nil {
			log.WithError(err).Errorf("failed to index has parent/child for `%d` from `%s`", dbUser.GitlabId, msg.Link)
			return err
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Errorf("failed to make add user transaction")
		return err
	}

	err = p.SendDeleteUser(msg.RequestId, dbUser)
	if err != nil {
		log.WithError(err).Errorf("failed to publish delete users")
		return err
	}

	log.Infof("finish handling add user message action")
	return nil
}

func (p *processor) addUser(username, link string, accessLevel int64, logger *logan.Entry) (*data.Permission, error) {
	typeTo, err := p.getLinkType(link, pqueue.NormalPriority)
	if err != nil {
		logger.WithError(err).Errorf("failed to get link type")
		return nil, err
	}

	userApi, err := gitlab.GetUser(p.pqueues.UserPQueue, any(p.gitlabClient.GetUserFromApi), []any{any(username)}, pqueue.NormalPriority)
	if err != nil {
		logger.WithError(err).Errorf("failed to get user from API")
		return nil, err
	}

	permission, err := gitlab.GetPermission(p.pqueues.SuperUserPQueue, any(p.gitlabClient.AddUsersFromApi), []any{
		any(link),
		any(typeTo),
		any(data.Permission{
			GitlabId:    userApi.GitlabId,
			AccessLevel: accessLevel,
			Link:        link,
			Type:        typeTo,
		}),
	}, pqueue.NormalPriority)
	if err != nil {
		logger.WithError(err).Errorf("failed to add user `%d` in `%s`", userApi.GitlabId, link)
		return nil, err
	}
	permission.Type = typeTo

	return permission, nil
}
