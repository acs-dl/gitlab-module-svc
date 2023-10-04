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

func (p *processor) validateRemoveUser(msg data.ModulePayload) error {
	return validation.Errors{
		"link":     validation.Validate(msg.Link, validation.Required),
		"username": validation.Validate(msg.Username, validation.Required),
	}.Filter()
}

func (p *processor) HandleRemoveUserAction(msg data.ModulePayload) error {
	log := p.log.WithField("message", msg.RequestId)
	log.Infof("start handling remove user action")

	err := p.validateRemoveUser(msg)
	if err != nil {
		log.WithError(err).Errorf("failed to validate fields")
		return errors.Wrap(err, "Request is not valid")
	}
	msg.Link = strings.ToLower(msg.Link)

	msg.Type, err = p.getLinkType(msg.Link, pqueue.NormalPriority)
	if err != nil {
		p.log.WithError(err).Errorf("failed to get link type", msg.RequestId)
		return errors.Wrap(err, "failed to get link type")
	}

	userApi, err := gitlab.GetUser(p.pqueues.UserPQueue, any(p.gitlabClient.GetUserFromApi), []any{any(msg.Username)}, pqueue.NormalPriority)
	if err != nil {
		log.WithError(err).Errorf("failed to get user from API")
		return err
	}

	dbUser, err := p.usersQ.FilterByUsernames(msg.Username).Get()
	if err != nil {
		log.WithError(err).Errorf("failed to get user by `%s` username", msg.Username)
		return errors.Errorf("Failed to get user `%s` by Gitlab username from database", msg.Username)
	}

	if dbUser == nil {
		log.Errorf("no user with such username `%s`", msg.Username)
		return errors.Errorf("No user with Gitlab username `%s` was found in database", msg.Username)
	}

	err = gitlab.GetRequestError(p.pqueues.SuperUserPQueue, any(p.gitlabClient.RemoveUserFromApi), []any{
		any(msg.Link),
		any(msg.Type),
		any(userApi.GitlabId),
	}, pqueue.NormalPriority)
	if err != nil {
		log.WithError(err).Errorf("failed to remove `%d` from `%s` via gitlab API", userApi.GitlabId, msg.Link)
		return err
	}

	err = p.managerQ.Transaction(func() error {
		err = p.deleteLowerLevelPermissions(userApi.GitlabId, msg.Link, msg.Type, log)
		if err != nil {
			log.WithError(err).Errorf("failed to delete lower level permissions")
			return err
		}

		permissions, err := p.permissionsQ.FilterByGitlabIds(userApi.GitlabId).Select()
		if err != nil {
			log.WithError(err).Errorf("failed to select permissions by gitlab id `%d`", userApi.GitlabId)
			return errors.Errorf("Failed to select permissions by `%s` Gitlab ID from database", userApi.GitlabId)
		}
		if len(permissions) == 0 {
			err = p.usersQ.FilterByGitlabIds(userApi.GitlabId).Delete()
			if err != nil {
				log.WithError(err).Errorf("failed to delete user by `%s` gitlab id", userApi.GitlabId)
				return errors.Errorf("Failed to delete user `%s` by Gitlab ID from database", userApi.GitlabId)
			}

			if dbUser.Id == nil {
				err = p.SendDeleteUser(msg.RequestId, *dbUser)
				if err != nil {
					log.WithError(err).Errorf("failed to publish delete user")
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Errorf("failed to make remove user transaction")
		return err
	}

	log.Infof("finish handling remove user message action")
	return nil
}

func (p *processor) deleteLowerLevelPermissions(gitlabId int64, link, typeTo string, logger *logan.Entry) error {
	err := p.permissionsQ.FilterByGitlabIds(gitlabId).FilterByTypes(typeTo).FilterByLinks(link).Delete()
	if err != nil {
		logger.WithError(err).Errorf("failed to delete permission for `%d` from `%s` from db", gitlabId, link)
		return errors.Errorf("Failed to delete permission for `%d` from `%s` from database", gitlabId, link)
	}

	permissions, err := p.permissionsQ.FilterByParentLinks(link).FilterByGitlabIds(gitlabId).Select()
	if err != nil {
		logger.WithError(err).Errorf("failed to select permission for `%d` by `%s` parent link from db", gitlabId, link)
		return errors.Errorf("Failed to select permission for `%d` by `%s` parent link from db", gitlabId, link)
	}

	if len(permissions) == 0 {
		return nil
	}

	for _, permission := range permissions {
		err = p.deleteLowerLevelPermissions(permission.GitlabId, permission.Link, permission.Type, logger)
		if err != nil {
			p.log.WithError(err).Errorf("failed to delete lower level permissions")
			return err
		}
	}

	return nil
}
