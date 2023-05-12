package handlers

import (
	"net/http"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"gitlab.com/distributed_lab/ape"
)

func GetUserRolesMap(w http.ResponseWriter, r *http.Request) {
	result := newModuleRolesResponse()

	result.Data.Attributes["super_admin"] = data.Roles[50]
	result.Data.Attributes["admin"] = data.Roles[40]
	result.Data.Attributes["user"] = data.Roles[10]

	ape.Render(w, result)
}
