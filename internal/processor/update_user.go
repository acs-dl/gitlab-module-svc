package processor

import (
	"strings"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateUpdateUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":         validation.Validate(msg.Link, validation.Required),
		"username":     validation.Validate(msg.Username, validation.Required),
		"access_level": validation.Validate(msg.AccessLevel, validation.Required),
	}.Filter()
}

func (p *processor) HandleUpdateUserAction(msg data.ModulePayload) error {
	p.log.Infof("start handle message action with id `%s`", msg.RequestId)

	err := p.validateUpdateUser(msg)
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

	user, err := p.checkUserExistence(msg.Username)
	if err != nil {
		p.log.WithError(err).Errorf("failed to check user existence for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to check user existence")
	}

	err = p.updateUser(data.Permission{
		GitlabId:    user.GitlabId,
		Username:    msg.Username,
		AccessLevel: msg.AccessLevel,
		Link:        msg.Link,
		Type:        msg.Type,
	})
	if err != nil {
		p.log.WithError(err).Errorf("failed to update user for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to update user")
	}

	err = p.indexHasParentChild(user.GitlabId, msg.Link)
	if err != nil {
		p.log.WithError(err).Errorf("failed to check has parent/child for message action with id `%s`", msg.RequestId)
		return errors.Wrap(err, "failed to check parent level")
	}

	p.log.Infof("finish handle message action with id `%s`", msg.RequestId)
	return nil
}

func (p *processor) updateUser(info data.Permission) error {
	permission, err := gitlab.GetPermission(
		p.pqueues.SuperUserPQueue,
		any(p.gitlabClient.UpdateUserFromApi),
		[]any{any(info)},
		pqueue.NormalPriority)
	if err != nil {
		return errors.Wrap(err, "some error while updating user from api")
	}

	if permission == nil {
		return errors.Errorf("something wrong with updating user from api")
	}

	err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(info.Link).
		Update(data.PermissionToUpdate{
			Username:    &permission.Username,
			AccessLevel: &permission.AccessLevel,
		})
	if err != nil {
		return errors.Wrap(err, "failed to update user in permission db")
	}
	return nil
}

func (p *processor) checkUserExistence(username string) (*data.User, error) {
	dbUser, err := p.usersQ.FilterByUsernames(username).Get()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user from user db")
	}

	if dbUser == nil {
		return nil, errors.New("no user with such username")
	}

	userApi, err := gitlab.GetUser(p.pqueues.UserPQueue, any(p.gitlabClient.GetUserFromApi), []any{any(username)}, pqueue.NormalPriority)
	if err != nil {
		return nil, errors.Wrap(err, "some error while getting user from api")
	}

	if userApi == nil {
		return nil, errors.Errorf("something wrong with user from api")
	}

	return dbUser, nil
}
