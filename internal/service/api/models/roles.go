package models

import "github.com/acs-dl/gitlab-module-svc/resources"

type RoleItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var roles = []resources.AccessLevel{
	{Name: "Guest", Value: 10},
	{Name: "Reporter", Value: 20},
	{Name: "Developer", Value: 30},
	{Name: "Maintainer", Value: 40},
	{Name: "Owner", Value: 50},
}

func NewRolesModel(found bool, roles []resources.AccessLevel) resources.Roles {
	result := resources.Roles{
		Key: resources.Key{
			ID:   "roles",
			Type: resources.ROLES,
		},
		Attributes: resources.RolesAttributes{
			Req:  found,
			List: roles,
		},
	}

	return result
}

func NewRolesResponse(found bool, startLevel, current int64) resources.RolesResponse {
	if !found {
		return resources.RolesResponse{
			Data: NewRolesModel(found, []resources.AccessLevel{}),
		}
	}

	newRoles := newRolesArray(startLevel, current)
	if len(newRoles) == 0 {
		return resources.RolesResponse{
			Data: NewRolesModel(!found, []resources.AccessLevel{}),
		}
	}

	return resources.RolesResponse{
		Data: NewRolesModel(found, newRoles),
	}
}

func newRolesArray(startLevel, current int64) []resources.AccessLevel {
	var result []resources.AccessLevel

	for _, role := range roles {
		if role.Value == current {
			continue
		}
		if role.Value > startLevel {
			result = append(result, role)
		}
	}

	return result
}
