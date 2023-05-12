package models

import (
	"strconv"
	"time"

	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/resources"
)

func NewUserPermissionModel(permission data.Sub, counter int) resources.UserPermission {
	var expiresAt *time.Time = nil
	if !permission.ExpiresAt.IsZero() {
		expiresAt = &permission.ExpiresAt
	}

	result := resources.UserPermission{
		Key: resources.Key{
			ID:   strconv.Itoa(counter),
			Type: resources.USER_PERMISSION,
		},
		Attributes: resources.UserPermissionAttributes{
			Username: permission.Username,
			ModuleId: permission.GitlabId,
			Path:     permission.Path,
			UserId:   permission.UserId,
			Type:     permission.Type,
			Link:     permission.Link,
			AccessLevel: resources.AccessLevel{
				Name:  data.Roles[permission.AccessLevel],
				Value: permission.AccessLevel,
			},
			Deployable: permission.HasChild,
			ExpiresAt:  expiresAt,
		},
	}

	return result
}

func NewUserPermissionList(permissions []data.Sub) []resources.UserPermission {
	result := make([]resources.UserPermission, len(permissions))
	for i, permission := range permissions {
		result[i] = NewUserPermissionModel(permission, i)
	}
	return result
}

func NewUserPermissionListResponse(permissions []data.Sub) UserPermissionListResponse {
	return UserPermissionListResponse{
		Data: NewUserPermissionList(permissions),
	}
}

type UserPermissionListResponse struct {
	Meta  Meta                       `json:"meta"`
	Data  []resources.UserPermission `json:"data"`
	Links *resources.Links           `json:"links"`
}
