package data

import "gitlab.com/distributed_lab/kit/pgdb"

type Subs interface {
	New() Subs

	Insert(sub Sub) error
	Delete() error
	Select() ([]Sub, error)
	Get() (*Sub, error)

	FilterByLinks(links ...string) Subs
	FilterByTypes(types ...string) Subs
	FilterByIds(ids ...int64) Subs
	SearchBy(search string) Subs

	WithPermissions() Subs
	FilterByGitlabIds(gitlabIds ...int64) Subs
	FilterByParentLinks(parentLinks ...string) Subs
	FilterByParentIds(parentIds ...int64) Subs
	FilterByHasParent(hasParent bool) Subs
	FilterByUserIds(userIds ...int64) Subs

	OrderBy(columns ...string) Subs

	Count() Subs
	CountWithPermissions() Subs
	GetTotalCount() (int64, error)

	Page(pageParams pgdb.OffsetPageParams) Subs
}

type Sub struct {
	Id          int64  `json:"id" db:"id" structs:"id"`
	Path        string `json:"path" db:"path" structs:"path"`
	Link        string `json:"link" db:"subs_link" structs:"link"`
	Type        string `json:"type" db:"subs_type" structs:"type"`
	ParentId    *int64 `json:"parent_id" db:"parent_id" structs:"parent_id"`
	*Permission `structs:",omitempty"`
}
