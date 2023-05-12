package postgres

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/fatih/structs"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	permissionsTableName         = "permissions"
	permissionsRequestIdColumn   = permissionsTableName + ".request_id"
	permissionsUserIdColumn      = permissionsTableName + ".user_id"
	permissionsUsernameColumn    = permissionsTableName + ".username"
	permissionsNameColumn        = permissionsTableName + ".name"
	permissionsGitlabIdColumn    = permissionsTableName + ".gitlab_id"
	permissionsLinkColumn        = permissionsTableName + ".link"
	permissionsAccessLevelColumn = permissionsTableName + ".access_level"
	permissionsTypeColumn        = permissionsTableName + ".type"
	permissionsCreatedAtColumn   = permissionsTableName + ".created_at"
	permissionsExpiresAtColumn   = permissionsTableName + ".expires_at"
	permissionsUpdatedAtColumn   = permissionsTableName + ".updated_at"
	permissionsParentLinkColumn  = permissionsTableName + ".parent_link"
	permissionsHasParentColumn   = permissionsTableName + ".has_parent"
	permissionsHasChildColumn    = permissionsTableName + ".has_child"
)

type PermissionsQ struct {
	db            *pgdb.DB
	selectBuilder sq.SelectBuilder
	deleteBuilder sq.DeleteBuilder
	updateBuilder sq.UpdateBuilder
}

var permissionsColumns = []string{
	permissionsRequestIdColumn,
	permissionsUserIdColumn,
	permissionsUsernameColumn,
	permissionsNameColumn,
	permissionsGitlabIdColumn,
	permissionsLinkColumn,
	permissionsAccessLevelColumn,
	permissionsTypeColumn,
	permissionsCreatedAtColumn,
	permissionsUpdatedAtColumn,
	permissionsExpiresAtColumn,
	permissionsHasParentColumn,
	permissionsHasChildColumn,
	permissionsParentLinkColumn,
}

func NewPermissionsQ(db *pgdb.DB) data.Permissions {
	return &PermissionsQ{
		db:            db.Clone(),
		selectBuilder: sq.Select(permissionsColumns...).From(permissionsTableName),
		deleteBuilder: sq.Delete(permissionsTableName),
		updateBuilder: sq.Update(permissionsTableName),
	}
}

func (q PermissionsQ) New() data.Permissions {
	return NewPermissionsQ(q.db)
}

func (q PermissionsQ) Update(permission data.PermissionToUpdate) error {
	updatedAt := time.Now()
	permission.UpdatedAt = &updatedAt

	q.updateBuilder = q.updateBuilder.SetMap(structs.Map(permission))

	return q.db.Exec(q.updateBuilder)
}

func (q PermissionsQ) Select() ([]data.Permission, error) {
	var result []data.Permission

	err := q.db.Select(&result, q.selectBuilder)

	return result, err
}

func (q PermissionsQ) Upsert(permission data.Permission) error {
	updateStmt, args := sq.Update(" ").
		Set("updated_at", time.Now()).
		Set("username", permission.Username).
		Set("name", permission.Name).
		Set("expires_at", permission.ExpiresAt).
		Set("access_level", permission.AccessLevel).MustSql()

	query := sq.Insert(permissionsTableName).SetMap(structs.Map(permission)).
		Suffix("ON CONFLICT (gitlab_id, link) DO "+updateStmt, args...)

	return q.db.Exec(query)
}

func (q PermissionsQ) Get() (*data.Permission, error) {
	var result data.Permission

	err := q.db.Get(&result, q.selectBuilder)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q PermissionsQ) Delete() error {
	var deleted []data.Permission

	err := q.db.Select(&deleted, q.deleteBuilder.Suffix("RETURNING *"))
	if err != nil {
		return err
	}

	if len(deleted) == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (q PermissionsQ) FilterByGitlabIds(ids ...int64) data.Permissions {
	equalGitlabIds := sq.Eq{permissionsGitlabIdColumn: ids}

	q.selectBuilder = q.selectBuilder.Where(equalGitlabIds)
	q.deleteBuilder = q.deleteBuilder.Where(equalGitlabIds)
	q.updateBuilder = q.updateBuilder.Where(equalGitlabIds)

	return q
}

func (q PermissionsQ) FilterByUsernames(usernames ...string) data.Permissions {
	equalUsernames := sq.Eq{permissionsUsernameColumn: usernames}

	q.selectBuilder = q.selectBuilder.Where(equalUsernames)
	q.deleteBuilder = q.deleteBuilder.Where(equalUsernames)
	q.updateBuilder = q.updateBuilder.Where(equalUsernames)

	return q
}

func (q PermissionsQ) FilterByLinks(links ...string) data.Permissions {
	equalLinks := sq.Eq{permissionsLinkColumn: links}

	q.selectBuilder = q.selectBuilder.Where(equalLinks)
	q.deleteBuilder = q.deleteBuilder.Where(equalLinks)
	q.updateBuilder = q.updateBuilder.Where(equalLinks)

	return q
}

func (q PermissionsQ) FilterByTypes(types ...string) data.Permissions {
	equalTypes := sq.Eq{permissionsTypeColumn: types}

	q.selectBuilder = q.selectBuilder.Where(equalTypes)
	q.deleteBuilder = q.deleteBuilder.Where(equalTypes)
	q.updateBuilder = q.updateBuilder.Where(equalTypes)

	return q
}

func (q PermissionsQ) FilterByParentLinks(parentLinks ...string) data.Permissions {
	equalParentLinks := sq.Eq{permissionsParentLinkColumn: parentLinks}
	if len(parentLinks) == 0 {
		equalParentLinks = sq.Eq{permissionsParentLinkColumn: nil}
	}

	q.selectBuilder = q.selectBuilder.Where(equalParentLinks)
	q.deleteBuilder = q.deleteBuilder.Where(equalParentLinks)
	q.updateBuilder = q.updateBuilder.Where(equalParentLinks)

	return q
}

func (q PermissionsQ) FilterByHasParent(hasParent bool) data.Permissions {
	equalHasParent := sq.Eq{permissionsHasParentColumn: hasParent}

	q.selectBuilder = q.selectBuilder.Where(equalHasParent)
	q.deleteBuilder = q.deleteBuilder.Where(equalHasParent)
	q.updateBuilder = q.updateBuilder.Where(equalHasParent)

	return q
}

func (q PermissionsQ) FilterByGreaterTime(time time.Time) data.Permissions {
	greaterTime := sq.Gt{permissionsUpdatedAtColumn: time}

	q.selectBuilder = q.selectBuilder.Where(greaterTime)
	q.deleteBuilder = q.deleteBuilder.Where(greaterTime)
	q.updateBuilder = q.updateBuilder.Where(greaterTime)

	return q
}

func (q PermissionsQ) FilterByLowerTime(time time.Time) data.Permissions {
	lowerTime := sq.Lt{permissionsUpdatedAtColumn: time}

	q.selectBuilder = q.selectBuilder.Where(lowerTime)
	q.deleteBuilder = q.deleteBuilder.Where(lowerTime)
	q.updateBuilder = q.updateBuilder.Where(lowerTime)

	return q
}
