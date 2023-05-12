package data

import (
	"time"

	"gitlab.com/distributed_lab/kit/pgdb"
)

type Users interface {
	New() Users

	Upsert(user User) error
	Delete() error
	Select() ([]User, error)
	Get() (*User, error)

	FilterByLowerTime(time time.Time) Users
	FilterById(id *int64) Users
	FilterByUsernames(usernames ...string) Users
	FilterByGitlabIds(gitlabIds ...int64) Users
	SearchBy(search string) Users

	Count() Users
	GetTotalCount() (int64, error)

	Page(pageParams pgdb.OffsetPageParams) Users
}

type User struct {
	Id             *int64    `json:"-" db:"id" structs:"id,omitempty"`
	GitlabUsername string    `json:"username" db:"username" structs:"username"`
	Name           string    `json:"name" db:"name" structs:"name"`
	GitlabId       int64     `json:"id" db:"gitlab_id" structs:"gitlab_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" structs:"-"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at" structs:"-"`
	AvatarUrl      string    `json:"avatar_url" db:"avatar_url" structs:"avatar_url"`
	Submodule      *string   `json:"-" db:"-" structs:"-"`
	AccessLevel    *string   `json:"-" db:"-" structs:"-"`
}

type UnverifiedUser struct {
	CreatedAt time.Time `json:"created_at"`
	Module    string    `json:"module"`
	Submodule string    `json:"submodule"`
	ModuleId  string    `json:"module_id"`
	Email     *string   `json:"email,omitempty"`
	Name      *string   `json:"name,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Username  *string   `json:"username,omitempty"`
}
