package keystore

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type Store interface {
	GetPublicKey(ctx context.Context, keyIRI string) (model.PublicKey, error)
	SavePublicKey(ctx context.Context, key *model.PublicKey) error
	SavePrivateKey(ctx context.Context, key *model.PrivateKey) error
}
