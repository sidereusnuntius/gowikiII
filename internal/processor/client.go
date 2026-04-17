package processor

import (
	"context"
	"database/sql"

	"github.com/riverqueue/river"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
)

type Client interface {
	Publish(ctx context.Context, job river.JobArgs) error
}

type RiverClient struct {
	Client *river.Client[*sql.Tx]
}

func (rc *RiverClient) Publish(ctx context.Context, job river.JobArgs) error {
	var err error
	tx, ok := txdb.GetTransaction(ctx)
	if ok {
		_, err = rc.Client.InsertTx(ctx, tx, job, nil)
	} else {
		_, err = rc.Client.Insert(ctx, job, nil)
	}
	return err
}
