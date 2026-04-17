package federation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
	"github.com/sidereusnuntius/gowiki/internal/processor"
	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

type Client interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

type ActorService interface {
	CacheRemoteActor(ctx context.Context, actor activitystreams.Actor) error
	ActorExists(ctx context.Context, id string) (bool, error)
	GetLocalActor(ctx context.Context, username string) (model.Actor, error)
}

type ArticleService interface {
	RemotePatch(ctx context.Context, patch activitystreams.Patch) error
	ArticleContent(ctx context.Context, req *model.ArticleRequest) (model.ArticleContent, error)
}

const (
	ActivityPatch = "Patch"
)

type Federation struct {
	Config    config.WikiConfig
	Actors    ActorService
	Articles  ArticleService
	Client    Client
	Store     Store
	Publisher processor.Client
}

func New(config config.WikiConfig, store Store, client Client, actors ActorService, articles ArticleService, publisher processor.Client) *Federation {
	return &Federation{
		Config:    config,
		Actors:    actors,
		Articles:  articles,
		Client:    client,
		Store:     store,
		Publisher: publisher,
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

// CheckOriginHost verifies whether an activity coming from the host is allowed. It will first check the local store;
// if there is no record for the host in the local store, it will create a new one, fetch the host's instance actor,
// and it might either reject or accept activities coming from unknown hosts, depending on the configuration defined.
func (f *Federation) CheckOriginHost(ctx context.Context, name string) (bool, error) {
	host, err := f.Store.GetHost(ctx, name)
	if err != nil {
		if !errors.Is(err, wikierr.ErrNotFound) {
			return false, err
		}

		host = model.Host{
			Host:   name,
			Status: model.HostUnknown,
		}

		if err = f.Store.SaveHost(ctx, &host); err != nil {
			return false, err
		}

		args := processor.FetchActorArgs{
			IRI:           "http://" + host.Host,
			InstanceActor: true,
		}

		// Fetch instance actor asynchronously.
		if err = f.Publisher.Publish(ctx, args); err != nil {
			return false, err
		}

		// TODO: decide whether the request should be allowed based on the wiki's config.
		return true, nil
	}

	return host.Status != model.Blocked, nil
}
