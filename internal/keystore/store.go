package keystore

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type Store interface {
	SavePublicKey(ctx context.Context, key *model.PublicKey) error
	SavePrivateKey(ctx context.Context, key *model.PrivateKey) error
}
