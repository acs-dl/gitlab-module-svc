package handlers

import (
	"net/http"
	"strings"

	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/models"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/requests"
	"github.com/acs-dl/gitlab-module-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func GetRoles(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetRolesRequest(r)
	if err != nil {
		Log(r).WithError(err).Errorf("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if request.Link == nil {
		Log(r).Warnf("no link was provided")
		ape.RenderErr(w, problems.NotFound())
		return
	}
	link := strings.ToLower(*request.Link)

	if request.Username == nil {
		Log(r).Warnf("no username was provided")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	permission, err := PermissionsQ(r).FilterByUsernames(*request.Username).FilterByLinks(link).Get()
	if err != nil {
		Log(r).WithError(err).Errorf("failed to get permission from `%s` to `%s`", link, *request.Username)
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if permission != nil {
		if permission.ParentLink == nil {
			Log(r).Warnf("no parent permissions => all roles")
			ape.Render(w, models.NewRolesResponse(true, 0, permission.AccessLevel))
			return
		}

		parentPermission, err := PermissionsQ(r).FilterByUsernames(*request.Username).FilterByLinks(*permission.ParentLink).Get()
		if err != nil {
			Log(r).WithError(err).Errorf("failed to get parent permission from `%s` to `%s`", *permission.ParentLink, *request.Username)
			ape.RenderErr(w, problems.BadRequest(err)...)
			return
		}

		ape.Render(w, models.NewRolesResponse(true, parentPermission.AccessLevel, permission.AccessLevel))
		return
	}

	response, err := checkRemotePermission(r, *request.Username, link)
	if err != nil {
		Log(r).WithError(err).Errorf("failed to check remote permission")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if response == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, response)
}

func checkRemotePermission(r *http.Request, username, link string) (*resources.RolesResponse, error) {
	pqs := pqueue.PQueuesInstance(ParentContext(r.Context()))
	gitlabClient := gitlab.GitlabClientInstance(ParentContext(r.Context()))

	typeSub, err := gitlab.GetPermissionWithType(pqs.SuperUserPQueue, any(gitlabClient.FindTypeFromApi), []any{any(link)}, pqueue.HighPriority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get type from api")
	}

	if typeSub == nil {
		Log(r).Warnf("no group/project `%s` was found", link)
		return nil, nil
	}

	userApi, err := gitlab.GetUser(pqs.UserPQueue, any(gitlabClient.GetUserFromApi), []any{any(username)}, pqueue.HighPriority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user from gitlab api")
	}

	if userApi == nil {
		Log(r).Warnf("no user `%s` was found", username)
		return nil, nil
	}

	permission, err := gitlab.GetPermission(pqs.SuperUserPQueue, any(gitlabClient.CheckUserFromApi), []any{
		any(link),
		any(typeSub.Type),
		any(userApi.GitlabId),
	}, pqueue.HighPriority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check user from api")
	}

	var response resources.RolesResponse
	if permission == nil {
		Log(r).Infof("no permission for user was found")
		response = models.NewRolesResponse(true, 0, 0)
		return &response, nil
	}

	if permission != nil && permission.ParentLink == nil {
		Log(r).Infof("no parent permissions, so got all roles")
		response = models.NewRolesResponse(true, 0, permission.AccessLevel)
		return &response, nil
	}

	parentPermission, err := PermissionsQ(r).FilterByUsernames(username).FilterByLinks(*permission.ParentLink).Get()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get parent permission")
	}

	if parentPermission == nil {
		return nil, errors.Wrap(err, "parent permission is empty")
	}

	response = models.NewRolesResponse(true, parentPermission.AccessLevel, permission.AccessLevel)
	return &response, nil
}
