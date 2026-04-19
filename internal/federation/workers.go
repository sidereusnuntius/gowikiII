package federation

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
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
	Client Client
	Store  Store
}

func (w *DiscoverInstanceWorker) Work(ctx context.Context, job *river.Job[DiscoverInstanceArgs]) error {
	host := job.Args.Host

	url := fmt.Sprintf("http://%s", host) // TODO: hardcoded HTTP.

	body, err := w.Client.Fetch(ctx, url)
	if wikierr.Is(err, wikierr.ErrNotFound) {
		wikilog.Logger.Debug().Msgf("host %s does not seem to expose an instance actor", host)
		return nil
	}

	actor, err := activitystreams.ReadObject(body)
	if err != nil {
		return err
	}

}
