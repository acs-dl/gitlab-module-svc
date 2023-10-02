package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/helpers"
	"github.com/fatih/structs"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	subsTableName      = "subs"
	subsIdColumn       = subsTableName + ".id"
	subsLinkColumn     = subsTableName + ".link"
	subsPathColumn     = subsTableName + ".path"
	subsTypeColumn     = subsTableName + ".type"
	subsParentIdColumn = subsTableName + ".parent_id"
)

type SubsQ struct {
	db            *pgdb.DB
	selectBuilder sq.SelectBuilder
	deleteBuilder sq.DeleteBuilder
}

var subsColumns = []string{
	subsIdColumn,
	subsLinkColumn + " as subs_link",
	subsPathColumn,
	subsTypeColumn + " as subs_type",
	subsParentIdColumn,
}

func NewSubsQ(db *pgdb.DB) data.Subs {
	return &SubsQ{
		db:            db.Clone(),
		selectBuilder: sq.Select(subsColumns...).From(subsTableName),
		deleteBuilder: sq.Delete(subsTableName),
	}
}

func (q SubsQ) New() data.Subs {
	return NewSubsQ(q.db)
}

func (q SubsQ) Insert(sub data.Sub) error {
	updateStmt, args := sq.Update(" ").
		Set("link", sub.Link).
		Set("path", sub.Path).
		MustSql()

	query := sq.Insert(subsTableName).SetMap(structs.Map(sub)).
		Suffix("ON CONFLICT (id) DO "+updateStmt, args...)

	return q.db.Exec(query)
}

func (q SubsQ) Select() ([]data.Sub, error) {
	var result []data.Sub

	err := q.db.Select(&result, q.selectBuilder)

	return result, err
}

func (q SubsQ) Get() (*data.Sub, error) {
	var result data.Sub

	err := q.db.Get(&result, q.selectBuilder)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q SubsQ) Delete() error {
	var deleted []data.Sub

	err := q.db.Select(&deleted, q.deleteBuilder.Suffix("RETURNING *"))
	if err != nil {
		return err
	}

	if len(deleted) == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (q SubsQ) WithPermissions() data.Subs {
	q.selectBuilder = sq.Select().
		Columns(helpers.RemoveDuplicateColumn(append(subsColumns, permissionsColumns...))...).
		From(subsTableName).
		LeftJoin(permissionsTableName + " ON " + permissionsLinkColumn + " = " + subsLinkColumn).
		Where(sq.NotEq{permissionsRequestIdColumn: nil})

	return q
}

func (q SubsQ) CountWithPermissions() data.Subs {
	q.selectBuilder = sq.Select("COUNT(*)").
		From(subsTableName).
		LeftJoin(permissionsTableName + " ON " + permissionsLinkColumn + " = " + subsLinkColumn).
		Where(sq.NotEq{permissionsRequestIdColumn: nil})

	return q
}

func (q SubsQ) FilterByParentLinks(parentLinks ...string) data.Subs {
	equalParentLinks := sq.Eq{permissionsParentLinkColumn: parentLinks}
	if len(parentLinks) == 0 {
		equalParentLinks = sq.Eq{permissionsParentLinkColumn: nil}
	}

	q.selectBuilder = q.selectBuilder.Where(equalParentLinks)
	q.deleteBuilder = q.deleteBuilder.Where(equalParentLinks)

	return q
}

func (q SubsQ) FilterByUserIds(userIds ...int64) data.Subs {
	equalUserIds := sq.Eq{permissionsUserIdColumn: userIds}

	if len(userIds) == 0 {
		equalUserIds = sq.Eq{permissionsUserIdColumn: nil}
	}

	q.selectBuilder = q.selectBuilder.Where(equalUserIds)
	q.deleteBuilder = q.deleteBuilder.Where(equalUserIds)

	return q
}

func (q SubsQ) FilterByGitlabIds(GitlabIds ...int64) data.Subs {
	equalGitlabIds := sq.Eq{permissionsGitlabIdColumn: GitlabIds}

	q.selectBuilder = q.selectBuilder.Where(equalGitlabIds)
	q.deleteBuilder = q.deleteBuilder.Where(equalGitlabIds)

	return q
}

func (q SubsQ) FilterByHasParent(hasParent bool) data.Subs {
	equalHasParent := sq.Eq{permissionsHasParentColumn: hasParent}

	q.selectBuilder = q.selectBuilder.Where(equalHasParent)
	q.deleteBuilder = q.deleteBuilder.Where(equalHasParent)

	return q
}

func (q SubsQ) SearchBy(search string) data.Subs {
	search = strings.Replace(search, " ", "%", -1)
	search = fmt.Sprint("%", search, "%")

	ilikeSearch := sq.ILike{subsPathColumn: search}

	q.selectBuilder = q.selectBuilder.Where(ilikeSearch)
	q.deleteBuilder = q.deleteBuilder.Where(ilikeSearch)

	return q
}

func (q SubsQ) FilterByLinks(links ...string) data.Subs {
	equalLinks := sq.Eq{subsLinkColumn: links}

	q.selectBuilder = q.selectBuilder.Where(equalLinks)
	q.deleteBuilder = q.deleteBuilder.Where(equalLinks)

	return q
}

func (q SubsQ) FilterByTypes(types ...string) data.Subs {
	equalTypes := sq.Eq{subsTypeColumn: types}

	q.selectBuilder = q.selectBuilder.Where(equalTypes)
	q.deleteBuilder = q.deleteBuilder.Where(equalTypes)

	return q
}

func (q SubsQ) FilterByIds(ids ...int64) data.Subs {
	equalIds := sq.Eq{subsIdColumn: ids}

	q.selectBuilder = q.selectBuilder.Where(equalIds)
	q.deleteBuilder = q.deleteBuilder.Where(equalIds)

	return q
}

func (q SubsQ) FilterByParentIds(parentIds ...int64) data.Subs {
	equalParentIds := sq.Eq{subsParentIdColumn: parentIds}
	if len(parentIds) == 0 {
		equalParentIds = sq.Eq{subsParentIdColumn: nil}
	}

	q.selectBuilder = q.selectBuilder.Where(equalParentIds)
	q.deleteBuilder = q.deleteBuilder.Where(equalParentIds)

	return q
}

func (q SubsQ) OrderBy(columns ...string) data.Subs {
	q.selectBuilder = q.selectBuilder.OrderBy(columns...)

	return q
}

func (q SubsQ) Count() data.Subs {
	q.selectBuilder = sq.Select("COUNT (*)").From(subsTableName)

	return q
}

func (q SubsQ) GetTotalCount() (int64, error) {
	var count int64

	err := q.db.Get(&count, q.selectBuilder)

	return count, err
}

func (q SubsQ) Page(pageParams pgdb.OffsetPageParams) data.Subs {
	q.selectBuilder = pageParams.ApplyTo(q.selectBuilder, subsTableName+".link")

	return q
}
