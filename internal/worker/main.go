package worker

import (
	"context"
	"fmt"
	"gitlab.com/distributed_lab/logan/v3"
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/config"
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/data/postgres"
	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	"github.com/acs-dl/gitlab-module-svc/internal/processor"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
)

const ServiceName = data.ModuleName + "-worker"

type IWorker interface {
	Run(ctx context.Context)
	ProcessPermissions(ctx context.Context) error
	RefreshSubmodules(msg data.ModulePayload) error
	GetEstimatedTime() time.Duration
}

type Worker struct {
	logger        *logan.Entry
	processor     processor.Processor
	gitlabClient  gitlab.GitlabClient
	linksQ        data.Links
	usersQ        data.Users
	subsQ         data.Subs
	permissionsQ  data.Permissions
	runnerDelay   time.Duration
	estimatedTime time.Duration
	pqueues       *pqueue.PQueues
}

func NewWorkerAsInterface(cfg config.Config, ctx context.Context) interface{} {
	return interface{}(&Worker{
		logger:        cfg.Log().WithField("runner", ServiceName),
		processor:     processor.ProcessorInstance(ctx),
		gitlabClient:  gitlab.GitlabClientInstance(ctx),
		pqueues:       pqueue.PQueuesInstance(ctx),
		linksQ:        postgres.NewLinksQ(cfg.DB()),
		subsQ:         postgres.NewSubsQ(cfg.DB()),
		usersQ:        postgres.NewUsersQ(cfg.DB()),
		permissionsQ:  postgres.NewPermissionsQ(cfg.DB()),
		estimatedTime: time.Duration(0),
		runnerDelay:   cfg.Runners().Worker,
	})
}

func (w *Worker) Run(ctx context.Context) {
	running.WithBackOff(
		ctx,
		w.logger,
		ServiceName,
		w.ProcessPermissions,
		w.runnerDelay,
		w.runnerDelay,
		w.runnerDelay,
	)
}

func (w *Worker) ProcessPermissions(_ context.Context) error {
	w.logger.Info("fetching links")

	startTime := time.Now()

	links, err := w.linksQ.Select()
	if err != nil {
		return errors.Wrap(err, "failed to get links")
	}

	reqAmount := len(links)
	if reqAmount == 0 {
		w.logger.Info("no links were found")
		return nil
	}

	w.logger.Infof("found %v links", reqAmount)

	for _, link := range links {
		w.logger.Infof("processing link `%s`", link.Link)

		err = w.createSubs(link.Link)
		if err != nil {
			w.logger.WithError(err).Errorf("failed to create subs for link `%s", link.Link)
			return errors.Wrap(err, "failed to create subs")
		}

		w.logger.WithField("link", link.Link).Info("link was processed successfully")

	}

	err = w.removeOldUsers(startTime)
	if err != nil {
		w.logger.WithError(err).Errorf("failed to remove old users")
		return errors.Wrap(err, "failed to remove old users")
	}

	err = w.removeOldPermissions(startTime)
	if err != nil {
		w.logger.WithError(err).Errorf("failed to remove old permissions")
		return errors.Wrap(err, "failed to remove old permissions")
	}

	w.estimatedTime = time.Now().Sub(startTime)
	return nil
}

func (w *Worker) removeOldUsers(borderTime time.Time) error {
	w.logger.Infof("started removing old users")

	users, err := w.usersQ.FilterByLowerTime(borderTime).Select()
	if err != nil {
		w.logger.Infof("failed to select users")
		return errors.Wrap(err, " failed to select users")
	}

	w.logger.Infof("found `%d` users to delete", len(users))

	for _, user := range users {
		if user.Id == nil { //if unverified user we need to remove them from `unverified-svc`
			err = w.processor.SendDeleteUser(uuid.New().String(), user)
			if err != nil {
				w.logger.WithError(err).Errorf("failed to publish delete user")
				return errors.Wrap(err, " failed to publish delete user")
			}
		}

		err = w.usersQ.FilterByGitlabIds(user.GitlabId).Delete()
		if err != nil {
			w.logger.Infof("failed to delete user with gitlab id `%d`", user.GitlabId)
			return errors.Wrap(err, " failed to delete user")
		}
	}

	w.logger.Infof("finished removing old users")
	return nil
}

func (w *Worker) removeOldPermissions(borderTime time.Time) error {
	w.logger.Infof("started removing old permissions")

	permissions, err := w.permissionsQ.FilterByLowerTime(borderTime).Select()
	if err != nil {
		w.logger.Infof("failed to select permissions")
		return errors.Wrap(err, " failed to select permissions")
	}

	w.logger.Infof("found `%d` permissions to delete", len(permissions))

	for _, permission := range permissions {
		err = w.permissionsQ.FilterByGitlabIds(permission.GitlabId).FilterByTypes(permission.Type).FilterByLinks(permission.Link).Delete()
		if err != nil {
			w.logger.Infof("failed to delete permission")
			return errors.Wrap(err, " failed to delete permission")
		}
	}

	w.logger.Infof("finished removing old permissions")
	return nil
}

func (w *Worker) createPermission(link string) error {
	w.logger.Infof("processing sub `%s`", link)

	if err := w.processor.HandleGetUsersAction(data.ModulePayload{
		RequestId: "from-worker",
		Link:      link,
	}); err != nil {
		w.logger.Errorf("failed to get users sub `%s`", link)
		return errors.Wrap(err, "failed to get users")
	}

	w.logger.Infof("successfully processed sub `%s", link)
	return nil
}

func (w *Worker) RefreshSubmodules(msg data.ModulePayload) error {
	w.logger.Infof("started refresh submodules")

	for _, link := range msg.Links {
		w.logger.Infof("started refreshing `%s`", link)

		err := w.createSubs(link)
		if err != nil {
			w.logger.Infof("failed to create subs for link `%s", link)
			return errors.Wrap(err, "failed to create subs")
		}
		w.logger.Infof("finished refreshing `%s`", link)
	}

	w.logger.Infof("finished refresh submodules")
	return nil
}

func (w *Worker) createSubs(link string) error {
	w.logger.Infof("creating subs for link `%s", link)

	checkType, err := gitlab.GetPermissionWithType(w.pqueues.SuperUserPQueue, any(w.gitlabClient.FindTypeFromApi), []any{any(link)}, pqueue.LowPriority)
	if err != nil {
		w.logger.Errorf("failed to get type for link `%s`", link)
		return errors.Wrap(err, "failed to get type")
	}

	if checkType == nil {
		w.logger.Errorf("failed to get sub for link `%s`", link)
		return errors.New("failed to get sub")
	}

	if validation.Validate(checkType.Type, validation.In(data.Group, data.Project)) != nil {
		return errors.New("something wrong with link type")
	}

	var sub = data.Sub{
		Id:   checkType.Sub.Id,
		Path: checkType.Sub.Path,
		Type: checkType.Type,
	}
	if checkType.Sub.ParentId != nil {
		sub.Link = link
		sub.ParentId = checkType.Sub.ParentId

	} else {
		sub.Link = checkType.Sub.Path
		sub.ParentId = nil
	}
	err = w.subsQ.Insert(sub)
	if err != nil {
		w.logger.Infof("failed to upsert sub for link `%s`", link)
		return errors.Wrap(err, "failed to upsert sub")
	}

	err = w.createPermission(link)
	if err != nil {
		w.logger.Infof("failed to create permissions for sub with link `%s`", link)
		return errors.Wrap(err, "failed to create permissions for sub")
	}

	if checkType.Type == data.Project {
		return nil
	}

	err = w.processNested(link, checkType.Sub.Id)
	if err != nil {
		w.logger.Infof("failed to index subs for link `%s`", link)
		return errors.Wrap(err, "failed to index subs")
	}

	w.logger.Infof("finished creating subs for link `%s", link)
	return nil
}

func (w *Worker) processNested(link string, parentId int64) error {
	w.logger.Debugf("processing link `%s`", link)

	projects, err := w.gitlabClient.GetProjectsFomApi(link)
	if err != nil {
		w.logger.Infof("failed to get projects for link `%s`", link)
		return errors.Wrap(err, fmt.Sprintf("failed to get projects for link `%s`", link))
	}

	for _, project := range projects {
		err = w.subsQ.Insert(data.Sub{
			Id:       project.Id,
			Path:     project.Path,
			Link:     link + "/" + project.Path,
			Type:     data.Project,
			ParentId: &parentId,
		})
		if err != nil {
			w.logger.Infof("failed to upsert sub with link `%s`", link+"/"+project.Path)
			return errors.Wrap(err, fmt.Sprintf("failed to get upsert sub with link `%s`", link+"/"+project.Path))
		}

		err = w.createPermission(link + "/" + project.Path)
		if err != nil {
			w.logger.Infof("failed to create permissions for sub with link `%s`", link+"/"+project.Path)
			return errors.Wrap(err, "failed to create permissions for sub")
		}
	}

	subgroups, err := w.gitlabClient.GetSubgroupsFomApi(link)
	if err != nil {
		w.logger.Infof("failed to get subgroups for link `%s`", link)
		return errors.Wrap(err, fmt.Sprintf("failed to get subgroups for link `%s`", link))
	}

	if len(subgroups) == 0 {
		return nil
	}

	for _, subgroup := range subgroups {
		err = w.subsQ.Insert(data.Sub{
			Id:       subgroup.Id,
			Path:     subgroup.Path,
			Link:     link + "/" + subgroup.Path,
			Type:     data.Group,
			ParentId: &parentId,
		})
		if err != nil {
			w.logger.Infof("failed to upsert sub with link `%s`", link+"/"+subgroup.Path)
			return errors.Wrap(err, fmt.Sprintf("failed to get upsert sub with link `%s`", link+"/"+subgroup.Path))
		}

		err = w.createPermission(link + "/" + subgroup.Path)
		if err != nil {
			w.logger.Infof("failed to create permissions for sub with link `%s`", link+"/"+subgroup.Path)
			return errors.Wrap(err, "failed to create permissions for sub")
		}

		err = w.processNested(link+"/"+subgroup.Path, subgroup.Id)
		if err != nil {
			w.logger.Infof("failed to make recursion")
			return errors.Wrap(err, "failed to make recursion")
		}
	}

	return nil
}

func (w *Worker) GetEstimatedTime() time.Duration {
	return w.estimatedTime
}
