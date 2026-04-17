package processor

import (
	"context"
	"database/sql"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riversqlite"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type Processor struct {
	client  *river.Client[*sql.Tx]
	workers *river.Workers
}

func (p *Processor) Client() Client {
	return &RiverClient{
		Client: p.client,
	}
}

func SqliteProcessor(ctx context.Context, db *sql.DB) (Processor, error) {
	wikilog.Logger.Debug().Msg("initializing Sqlite River queue")
	workers := river.NewWorkers()

	driver := riversqlite.New(db)

	migrator, err := rivermigrate.New(driver, nil)
	if err != nil {
		return Processor{}, err
	}

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return Processor{}, err
	}

	client, err := river.NewClient(driver, &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		return Processor{}, err
	}

	return Processor{
		client:  client,
		workers: workers,
	}, nil
}
