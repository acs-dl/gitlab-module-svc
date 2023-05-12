package handlers

import (
	"net/http"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/models"
	"github.com/acs-dl/gitlab-module-svc/internal/service/api/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetPermissions(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGetPermissionsRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("bad request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	var userIds []int64
	if request.UserId != nil {
		userIds = append(userIds, *request.UserId)
	}

	var parentLinks []string
	if request.ParentLink != nil {
		parentLinks = append(parentLinks, *request.ParentLink)
	}

	statement := SubsQ(r).WithPermissions().FilterByUserIds(userIds...).
		FilterByHasParent(false).FilterByParentLinks(parentLinks...)

	totalCount := SubsQ(r).CountWithPermissions().FilterByUserIds(userIds...).
		FilterByHasParent(false).FilterByParentLinks(parentLinks...)

	if request.Link != nil {
		statement = SubsQ(r).WithPermissions().FilterByUserIds(userIds...).
			SearchBy(*request.Link)
		totalCount = SubsQ(r).CountWithPermissions().FilterByUserIds(userIds...).
			SearchBy(*request.Link)
	}

	permissions, err := statement.Page(request.OffsetPageParams).Select()
	if err != nil {
		Log(r).WithError(err).Error("failed to get permissions")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	amount, err := totalCount.GetTotalCount()
	if err != nil {
		Log(r).WithError(err).Error("failed to get total count")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := models.NewUserPermissionListResponse(permissions)
	response.Meta.TotalCount = amount
	response.Links = data.GetOffsetLinksForPGParams(r, request.OffsetPageParams)

	ape.Render(w, response)
}
