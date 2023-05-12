package data

import "time"

type Permissions interface {
	New() Permissions

	Upsert(permission Permission) error
	Update(permission PermissionToUpdate) error
	Delete() error
	Select() ([]Permission, error)
	Get() (*Permission, error)

	FilterByGitlabIds(gitlabIds ...int64) Permissions
	FilterByUsernames(usernames ...string) Permissions
	FilterByTypes(types ...string) Permissions
	FilterByLinks(links ...string) Permissions
	FilterByGreaterTime(time time.Time) Permissions
	FilterByLowerTime(time time.Time) Permissions
	FilterByParentLinks(parentLinks ...string) Permissions
	FilterByHasParent(hasParent bool) Permissions
}

type Permission struct {
	RequestId     string    `json:"request_id" db:"request_id" structs:"request_id"`
	UserId        *int64    `json:"user_id" db:"user_id" structs:"user_id"`
	Name          string    `json:"name" db:"name" structs:"name"`
	Username      string    `json:"username" db:"username" structs:"username"`
	GitlabId      int64     `json:"id" db:"gitlab_id" structs:"gitlab_id"`
	AccessLevel   int64     `json:"access_level" db:"access_level" structs:"access_level"`
	HasParent     bool      `json:"-" db:"has_parent" structs:"-"`
	HasChild      bool      `json:"-" db:"has_child" structs:"-"`
	Link          string    `json:"link" db:"link" structs:"link"`
	ParentLink    *string   `json:"-" db:"parent_link" structs:"parent_link"`
	Type          string    `json:"type" db:"type" structs:"type"`
	CreatedAt     time.Time `json:"created_at" db:"created_at" structs:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at" structs:"-"`
	ExpiresAt     time.Time `json:"-" db:"expires_at" structs:"expires_at"`
	ExpiresString string    `json:"expires_at" db:"-" structs:"-"`
	AvatarUrl     string    `json:"avatar_url" db:"-" structs:"-"`
}

type PermissionToUpdate struct {
	Username    *string    `structs:"username,omitempty"`
	AccessLevel *int64     `structs:"access_level,omitempty"`
	UserId      *int64     `structs:"user_id,omitempty"`
	ParentLink  *string    `structs:"parent_link,omitempty"`
	HasParent   *bool      `structs:"has_parent,omitempty"`
	HasChild    *bool      `structs:"has_child,omitempty"`
	UpdatedAt   *time.Time `structs:"updated_at,omitempty"`
}
