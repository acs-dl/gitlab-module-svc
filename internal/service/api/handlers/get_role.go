package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/models"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetRole(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetRoleRequest(r)
	if err != nil {
		Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if request.AccessLevel == nil {
		Log(r).Errorf("no access level was provided")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	accessLevel, err := strconv.ParseInt(*request.AccessLevel, 10, 64)
	if err != nil {
		Log(r).WithError(err).Infof("failed to parse access_level `%s`", *request.AccessLevel)
		ape.RenderErr(w, problems.NotFound())
		return
	}

	name := data.Roles[accessLevel]
	if name == "" {
		Log(r).Errorf("no such access level `%s`", *request.AccessLevel)
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, models.NewRoleResponse(data.Roles[accessLevel], fmt.Sprintf("%d", accessLevel)))
}
