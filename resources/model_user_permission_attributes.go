/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import (
	"time"
)

type UserPermissionAttributes struct {
	AccessLevel AccessLevel `json:"access_level"`
	// indicates whether element have nested object
	Deployable bool `json:"deployable"`
	// shows when permission is expired
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// full path to repo for which was given access
	Link string `json:"link"`
	// user id from gitlab
	ModuleId int64 `json:"module_id"`
	// path to repo for which was given access
	Path string `json:"path"`
	// type of link for which was given access (group or project)
	Type string `json:"type"`
	// user id from identity
	UserId *int64 `json:"user_id,omitempty"`
	// username from gitlab
	Username string `json:"username"`
}
