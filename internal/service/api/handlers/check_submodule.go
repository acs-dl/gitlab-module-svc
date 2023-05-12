package handlers

import (
	"net/http"
	"strings"

	"github.com/acs-dl/gitlab-module-svc/internal/service/api/models"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func CheckSubmodule(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewCheckSubmoduleRequest(r)
	if err != nil {
		Log(r).WithError(err).Info("wrong request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if request.Link == nil {
		Log(r).Errorf("no link was provided")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	link := strings.ToLower(*request.Link)
	sub, err := SubsQ(r).FilterByLinks(link).Get()
	if err != nil {
		Log(r).WithError(err).Errorf("failed to get link `%s`", link)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if sub != nil {
		ape.Render(w, models.NewLinkResponse(sub.Path, true))
		return
	}

	Log(r).Warnf("no group/project was found")
	ape.Render(w, models.NewLinkResponse("", false))
}
