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
	Workers *river.Workers
}

func Register[T river.JobArgs](processor *Processor, worker river.Worker[T]) {
	river.AddWorker(processor.Workers, worker)
}

func SqliteProcessor(ctx context.Context, db *sql.DB) (*Processor, error) {
	wikilog.Logger.Debug().Msg("initializing Sqlite River queue")
	workers := river.NewWorkers()

	driver := riversqlite.New(db)

	migrator, err := rivermigrate.New(driver, nil)
	if err != nil {
		return nil, err
	}

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return nil, err
	}

	client, err := river.NewClient(driver, &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, err
	}

	return &Processor{
		client:  client,
		Workers: workers,
	}, nil
}

func (p *Processor) Client() Client {
	return &RiverClient{
		Client: p.client,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	return p.client.Start(ctx)
}

func (p *Processor) Stop(ctx context.Context) error {
	return p.client.Stop(ctx)
}
