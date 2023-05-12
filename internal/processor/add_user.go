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
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateAddUser(msg)
	if err != nil {
		p.log.WithError(err).Errorf("failed to validate fields for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to validate fields")
	}

	msg.Link = strings.ToLower(msg.Link)
	userId, err := strconv.ParseInt(msg.UserId, 10, 64)
	if err != nil {
		p.log.WithError(err).Errorf("failed to parse user id `%s` for message action with id `%s`", msg.UserId, msg.RequestId)
		return errors.Wrap(err, "failed to parse user id")
	}

	permission, err := p.addUser(msg.Username, msg.Link, msg.AccessLevel)
	if err != nil {
		p.log.WithError(err).Errorf("failed to add user from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while adding user from api")
	}

	permission.UserId = &userId
	permission.Link = msg.Link
	permission.RequestId = msg.RequestId
	permission.CreatedAt = time.Now()

	permission.ExpiresAt, err = helpers.ParseTime(permission.ExpiresString)
	if err != nil {
		p.log.WithError(err).Errorf("failed to parse time from API for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "some error while parsing time from api")
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
			p.log.WithError(err).Errorf("failed to creat user in user db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to create user in user db")
		}

		if err = p.permissionsQ.Upsert(*permission); err != nil {
			p.log.WithError(err).Errorf("failed to create permission in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to create permission in permission db")
		}

		//in case if we have some rows without id from identity
		err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).Update(data.PermissionToUpdate{UserId: permission.UserId})
		if err != nil {
			p.log.WithError(err).Errorf("failed to update user id in permission db for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to update user id in user db")
		}

		err = p.indexHasParentChild(dbUser.GitlabId, msg.Link)
		if err != nil {
			p.log.WithError(err).Errorf("failed to check has parent/child for message action with id `%s`", msg.RequestId)
			return errors.Wrap(err, "failed to check parent level")
		}

		return nil
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to make add user transaction for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to make add user transaction")
	}

	err = p.SendDeleteUser(msg.RequestId, dbUser)
	if err != nil {
		p.log.WithError(err).Errorf("failed to publish users for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to publish users")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}

func (p *processor) addUser(username, link string, accessLevel int64) (*data.Permission, error) {
	typeTo, err := p.getLinkType(link, pqueue.NormalPriority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get link type")
	}

	userApi, err := gitlab.GetUser(p.pqueues.UserPQueue, any(p.gitlabClient.GetUserFromApi), []any{any(username)}, pqueue.NormalPriority)
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting user id from api")
	}
	if userApi == nil {
		return nil, errors.New("no user was found from api")
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
		return nil, errors.Wrap(err, "some error while adding user from api")
	}
	permission.Type = typeTo

	return permission, nil
}
