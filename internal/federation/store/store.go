package hostsql

import (
	"context"
	"database/sql"

	sqlhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/sql"
	"github.com/sidereusnuntius/gowiki/internal/model"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

const (
	selectHost = `SELECT id, host, status, is_wiki, instance_actor_id FROM hosts WHERE host = ? LIMIT 1`
	insertHost = `INSERT INTO hosts (
		host,
		status,
		is_wiki,
		instance_actor_id
	) VALUES (?, ?, ?, ?) RETURNING id`
	updateHost = `UPDATE hosts SET
		host = ?,
		status = ?,
		is_wiki = ?,
		instance_actor_id = ?
		WHERE id = ?`
	updateHostStatus = `UPDATE hosts SET status = ? WHERE id = ?`
)

type HostsStore struct {
	DB *sql.DB
}

func New(db *sql.DB) *HostsStore {
	return &HostsStore{
		DB: db,
	}
}

func (hs *HostsStore) GetHost(ctx context.Context, hostname string) (model.Host, error) {
	row := txdb.GetExecutor(ctx, hs.DB).QueryRowContext(ctx, selectHost, hostname)

	var (
		host    model.Host
		status  int
		actorId sql.NullInt64
	)
	err := row.Scan(
		&host.ID,
		&status,
		&host.IsWiki,
		&actorId,
	)
	if err != nil {
		return model.Host{}, sqlhelpers.HandleErr(err)
	}

	if actorId.Valid {
		host.ActorID = actorId.Int64
	}
	host.Status = model.HostStatus(status)

	return host, nil
}

func (hs *HostsStore) SaveHost(ctx context.Context, host *model.Host) error {
	var (
		result sql.Result
		err    error
	)
	if host.ID != 0 {
		result, err = txdb.GetExecutor(ctx, hs.DB).ExecContext(ctx, updateHost,
			host.Host,
			int(host.Status),
			host.IsWiki,
			sqlhelpers.NullableInt64(host.ActorID),
			host.ID,
		)
	} else {
		result, err = txdb.GetExecutor(ctx, hs.DB).ExecContext(ctx, insertHost,
			host.Host,
			int(host.Status),
			host.IsWiki,
			sqlhelpers.NullableInt64(host.ActorID),
		)
	}

	if err != nil {
		return sqlhelpers.HandleErr(err)
	}

	if host.ID == 0 {
		host.ID, err = result.LastInsertId()
		if err != nil {
			return sqlhelpers.HandleErr(err)
		}
	}

	return nil
}

func (hs *HostsStore) UpdateHostStatus(ctx context.Context, id int64, status model.HostStatus) error {
	_, err := txdb.GetExecutor(ctx, hs.DB).ExecContext(ctx, updateHostStatus, int(status), id)
	return sqlhelpers.HandleErr(err)
}
