package federation

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
	"github.com/sidereusnuntius/gowiki/internal/processor"
	txdb "github.com/sidereusnuntius/gowiki/internal/transactions"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type DiscoverInstanceArgs struct {
	Host string `json:"host"`
}

func (DiscoverInstanceArgs) Kind() string {
	return "discover_instance"
}

type DiscoverInstanceWorker struct {
	river.WorkerDefaults[DiscoverInstanceArgs]
	Config    config.WikiConfig
	Client    Client
	Store     Store
	Actors    ActorService
	TxManager *txdb.TxManager
}

func (f *Federation) RegisterWorkers(p *processor.Processor) {
	discoverInstanceWorker := NewDiscoverInstanceWorker(f.Config, f.Store, f.TxManager, f.Actors, f.Client)

	processor.Register(p, discoverInstanceWorker)
}

func NewDiscoverInstanceWorker(config config.WikiConfig, store Store, manager *txdb.TxManager, actors ActorService, client Client) *DiscoverInstanceWorker {
	return &DiscoverInstanceWorker{
		Config:    config,
		Client:    client,
		Store:     store,
		Actors:    actors,
		TxManager: manager,
	}
}

func (w *DiscoverInstanceWorker) Work(ctx context.Context, job *river.Job[DiscoverInstanceArgs]) error {
	hostname := job.Args.Host
	wikilog.Logger.Debug().Str("host", hostname).Msg("attempting to fetch new host's instance actor")

	url := fmt.Sprintf("http://%s", hostname) // TODO: hardcoded HTTP.

	body, err := w.Client.Fetch(ctx, url)
	if wikierr.Is(err, wikierr.ErrNotFound) {
		wikilog.Logger.Debug().Msgf("host %s does not seem to expose an instance actor", hostname)
		return nil
	}

	object, err := activitystreams.ReadObject(body)
	if err != nil {
		return err
	}

	actor, err := object.AsActor()
	if err != nil {
		return err
	}

	host, err := w.Store.GetHost(ctx, hostname)
	if err != nil {
		if !wikierr.Is(err, wikierr.ErrNotFound) {
			return err
		}

		host = model.Host{
			Host: hostname,
		}
	}

	return w.TxManager.RunInTx(ctx, func(ctx context.Context) error {
		actorInternal, err := w.Actors.CacheRemoteActor(ctx, actor)
		if err != nil {
			return err
		}

		host.ActorID = actorInternal.ID
		host.Status = model.Fetched
		host.IsWiki = actor.Type == "Wiki"
		if err = w.Store.SaveHost(ctx, &host); err != nil {
			return err
		}

		follow := model.Follow{
			FollowerIRI: w.Config.URL.String(),
			FolloweeIRI: actorInternal.URI,
		}
		return w.Actors.Follow(ctx, &follow)
	})
}
