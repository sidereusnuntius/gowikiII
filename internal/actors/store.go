package actors

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type Store interface {
	SaveActor(ctx context.Context, actor *model.Actor) error
	// CreateSharedInboxIfNotExists checks if a shared inbox with the provided URI
	// exists and, if not, persists in the database. In both cases t returns the id
	//of the inserted shared inbox.
	CreateSharedInboxIfNotExists(ctx context.Context, uri string) (int64, error)
	GetActorByHandle(ctx context.Context, username, host string) (model.Actor, error)
	GetActorByIRI(ctx context.Context, iri string) (model.Actor, error)
	ActorExists(ctx context.Context, iri string) (bool, error)
}
