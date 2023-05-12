package handlers

import (
	"net/http"

	"github.com/acs-dl/gitlab-module-svc/internal/gitlab"
	"github.com/acs-dl/gitlab-module-svc/internal/pqueue"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/models"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetUsersRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("bad request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	username := ""
	if request.Username != nil {
		username = *request.Username
	}

	users, err := UsersQ(r).SearchBy(username).Select()
	if err != nil {
		Log(r).WithError(err).Errorf("failed to select users from db")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if len(users) != 0 {
		ape.Render(w, models.NewUserInfoListResponse(users, 0))
		return
	}

	users, err = gitlab.GetUsers(
		pqueue.PQueuesInstance(ParentContext(r.Context())).UserPQueue,
		any(gitlab.GitlabClientInstance(ParentContext(r.Context())).SearchByFromApi),
		[]any{any(username)}, pqueue.HighPriority)
	if err != nil {
		Log(r).WithError(err).Infof("failed to get users from api by `%s`", username)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, models.NewUserInfoListResponse(users, 0))
}
