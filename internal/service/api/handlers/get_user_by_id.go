package handlers

import (
	"net/http"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/models"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetUserById(w http.ResponseWriter, r *http.Request) {
	userId, err := requests.NewGetUserByIdRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("bad request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	user, err := UsersQ(r).FilterById(&userId).Get()
	if err != nil {
		Log(r).WithError(err).Errorf("failed to get user with id `%d`", userId)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if user == nil {
		Log(r).Errorf("no user with id `%d`", userId)
		ape.RenderErr(w, problems.NotFound())
		return
	}

	permission, err := PermissionsQ(r).FilterByGitlabIds(user.GitlabId).
		FilterByHasParent(false).FilterByParentLinks([]string{}...).Get()
	if err != nil {
		Log(r).Errorf("failed to get submodule for user user with id `%d`", userId)
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if permission != nil {
		accessLevel := data.Roles[permission.AccessLevel]
		user.Submodule = &permission.Link
		user.AccessLevel = &accessLevel
	}

	ape.Render(w, models.NewUserResponse(*user))
}
