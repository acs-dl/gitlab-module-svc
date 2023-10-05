package processor

import (
	"strings"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (p *processor) validateUpdateUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":         validation.Validate(msg.Link, validation.Required),
		"username":     validation.Validate(msg.Username, validation.Required),
		"access_level": validation.Validate(msg.AccessLevel, validation.Required),
	}.Filter()
}

// HandleUpdateUserAction .Wrap from return err was removed because
// errors must be human-readable from very low level to send them in FE.
// log must be put before every error to track it if any
func (p *processor) HandleUpdateUserAction(msg data.ModulePayload) error {
	log := p.log.WithField("message", msg.RequestId)
	log.Infof("start handling verify user action")

	err := p.validateUpdateUser(msg)
	if err != nil {
		log.WithError(err).Errorf("failed to validate fields")
		return errors.Wrap(err, "Request is not valid")
	}
	msg.Link = strings.ToLower(msg.Link)

	msg.Type, err = p.getLinkType(msg.Link, pqueue.NormalPriority)
	if err != nil {
		log.WithError(err).Errorf("failed to get link type")
		return err
	}

	user, err := p.checkUserExistence(msg.Username, log)
	if err != nil {
		log.WithError(err).Errorf("failed to check user existence")
		return err
	}

	err = p.updateUser(data.Permission{
		GitlabId:    user.GitlabId,
		Username:    msg.Username,
		AccessLevel: msg.AccessLevel,
		Link:        msg.Link,
		Type:        msg.Type,
	}, log)
	if err != nil {
		p.log.WithError(err).Errorf("failed to update user")
		return err
	}

	err = p.indexHasParentChild(user.GitlabId, msg.Link, log)
	if err != nil {
		log.WithError(err).Errorf("failed to index has parent/child for `%d` from `%s`", user.GitlabId, msg.Link)
		return err
	}

	log.Infof("finish handling update user message action")
	return nil
}

func (p *processor) updateUser(info data.Permission, logger *logan.Entry) error {
	permission, err := gitlab.GetPermission(
		p.pqueues.SuperUserPQueue,
		any(p.gitlabClient.UpdateUserFromApi),
		[]any{any(info)},
		pqueue.NormalPriority)
	if err != nil {
		logger.WithError(err).Errorf("failed to update user from api")
		return err
	}

	err = p.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByLinks(info.Link).
		Update(data.PermissionToUpdate{
			Username:    &permission.Username,
			AccessLevel: &permission.AccessLevel,
		})
	if err != nil {
		logger.WithError(err).Errorf("failed to update permission `%s` in `%s` in user db", info.Username, info.Link)
		return errors.Errorf("Failed to update permission for `%s` in `%s` in user database", info.Username, info.Link)
	}

	return nil
}

func (p *processor) checkUserExistence(username string, logger *logan.Entry) (*data.User, error) {
	dbUser, err := p.usersQ.FilterByUsernames(username).Get()
	if err != nil {
		logger.WithError(err).Errorf("failed to get user `%s` from user db", username)
		return nil, errors.Errorf("Failed to get user `%s` from user database", username)
	}

	if dbUser == nil {
		return nil, errors.Errorf("No user was found with `%s` username in database", username)
	}

	userApi, err := gitlab.GetUser(p.pqueues.UserPQueue, any(p.gitlabClient.GetUserFromApi), []any{any(username)}, pqueue.NormalPriority)
	if err != nil {
		logger.WithError(err).Errorf("failed to get user from API")
		return nil, err
	}

	if userApi == nil {
		return nil, errors.Errorf("No user was found with `%s` username in Gitlab API", username)
	}

	return dbUser, nil
}
