package manager

import (
	"github.com/acs-dl/gitlab-module-svc/internal/data"
	"github.com/acs-dl/gitlab-module-svc/internal/data/postgres"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Manager struct {
	db *pgdb.DB

	responses   data.Responses
	permissions data.Permissions
	users       data.Users
}

func NewManager(db *pgdb.DB) *Manager {
	return &Manager{
		db:          db,
		responses:   postgres.NewResponsesQ(db),
		permissions: postgres.NewPermissionsQ(db),
		users:       postgres.NewUsersQ(db),
	}
}

func (m *Manager) Transaction(fn func() error) error {
	return m.db.Transaction(fn)
}
