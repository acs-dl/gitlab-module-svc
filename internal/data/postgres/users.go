package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/fatih/structs"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	usersTableName       = "users"
	usersIdColumn        = usersTableName + ".id"
	usersUsernameColumn  = usersTableName + ".username"
	usersGitlabIdColumn  = usersTableName + ".gitlab_id"
	usersUpdatedAtColumn = usersTableName + ".updated_at"
)

type UsersQ struct {
	db            *pgdb.DB
	selectBuilder sq.SelectBuilder
	deleteBuilder sq.DeleteBuilder
}

func NewUsersQ(db *pgdb.DB) data.Users {
	return &UsersQ{
		db:            db.Clone(),
		selectBuilder: sq.Select("*").From(usersTableName),
		deleteBuilder: sq.Delete(usersTableName),
	}
}

func (q UsersQ) New() data.Users {
	return NewUsersQ(q.db)
}

func (q UsersQ) Upsert(user data.User) error {
	clauses := structs.Map(user)

	updateQuery := sq.Update(" ").
		Set("username", user.GitlabUsername).
		Set("name", user.Name).
		Set("avatar_url", user.AvatarUrl).
		Set("updated_at", time.Now())

	if user.Id != nil {
		updateQuery = updateQuery.Set("id", *user.Id)
	}

	updateStmt, args := updateQuery.MustSql()

	query := sq.Insert(usersTableName).SetMap(clauses).Suffix("ON CONFLICT (gitlab_id) DO "+updateStmt, args...)

	return q.db.Exec(query)
}

func (q UsersQ) Delete() error {
	var deleted []data.User

	err := q.db.Select(&deleted, q.deleteBuilder.Suffix("RETURNING *"))
	if err != nil {
		return err
	}

	if len(deleted) == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (q UsersQ) FilterById(id *int64) data.Users {
	equalId := sq.Eq{usersIdColumn: id}
	if id != nil {
		equalId = sq.Eq{usersIdColumn: *id}
	}

	q.selectBuilder = q.selectBuilder.Where(equalId)
	q.deleteBuilder = q.deleteBuilder.Where(equalId)

	return q
}

func (q UsersQ) FilterByGitlabIds(GitlabIds ...int64) data.Users {
	equalGitlabIds := sq.Eq{usersGitlabIdColumn: GitlabIds}

	q.selectBuilder = q.selectBuilder.Where(equalGitlabIds)
	q.deleteBuilder = q.deleteBuilder.Where(equalGitlabIds)

	return q
}

func (q UsersQ) FilterByUsernames(usernames ...string) data.Users {
	equalUsernames := sq.Eq{usersUsernameColumn: usernames}

	q.selectBuilder = q.selectBuilder.Where(equalUsernames)
	q.deleteBuilder = q.deleteBuilder.Where(equalUsernames)

	return q
}

func (q UsersQ) Get() (*data.User, error) {
	var result data.User

	err := q.db.Get(&result, q.selectBuilder)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q UsersQ) Select() ([]data.User, error) {
	var result []data.User

	err := q.db.Select(&result, q.selectBuilder)

	return result, err
}

func (q UsersQ) Page(pageParams pgdb.OffsetPageParams) data.Users {
	q.selectBuilder = pageParams.ApplyTo(q.selectBuilder, "username")

	return q
}

func (q UsersQ) SearchBy(search string) data.Users {
	search = strings.Replace(search, " ", "%", -1)
	search = fmt.Sprint("%", search, "%")

	q.selectBuilder = q.selectBuilder.Where(sq.ILike{"username": search})
	return q
}

func (q UsersQ) Count() data.Users {
	q.selectBuilder = sq.Select("COUNT (*)").From(usersTableName)

	return q
}

func (q UsersQ) GetTotalCount() (int64, error) {
	var count int64
	err := q.db.Get(&count, q.selectBuilder)

	return count, err
}

func (q UsersQ) FilterByLowerTime(time time.Time) data.Users {
	lowerTime := sq.Lt{usersUpdatedAtColumn: time}

	q.selectBuilder = q.selectBuilder.Where(lowerTime)
	q.deleteBuilder = q.deleteBuilder.Where(lowerTime)

	return q
}
