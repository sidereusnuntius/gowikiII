package federation

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
)

type Client interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

type ActorService interface {
	CacheRemoteActor(ctx context.Context, actor activitystreams.Actor) error
	ActorExists(ctx context.Context, id string) (bool, error)
}

type ArticleService interface {
	RemotePatch(ctx context.Context, patch activitystreams.Patch) error
}

const (
	ActivityPatch = "Patch"
)

type Federation struct {
	Config   config.WikiConfig
	Actors   ActorService
	Articles ArticleService
	Client   Client
}

func New(config config.WikiConfig, client Client, actors ActorService, articles ArticleService) Federation {
	return Federation{
		Config:   config,
		Actors:   actors,
		Articles: articles,
		Client:   client,
	}
}

func (f *Federation) InstanceInbox(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll(): %w", err)
	}

	activity, err := activitystreams.ReadActivity(body)
	if err != nil {
		return err
	}

	return f.ProcessActivity(r.Context(), activity)
}

func (f *Federation) EnsureActor(ctx context.Context, actorId string) error {
	cached, err := f.Actors.ActorExists(ctx, actorId)
	if err != nil {
		return err
	}

	// Actor already exists in local store.
	if cached {
		return nil
	}

	// Otherwise, fetch it from remote host.
	raw, err := f.Client.Fetch(ctx, actorId)
	if err != nil {
		return err
	}

	obj, err := activitystreams.ReadObject(raw)
	if err != nil {
		return err
	}

	actor, err := obj.AsActor()
	if err != nil {
		return err
	}

	return f.Actors.CacheRemoteActor(ctx, actor)
}
