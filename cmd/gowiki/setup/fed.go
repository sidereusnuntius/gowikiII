package setup

import (
	"github.com/sidereusnuntius/gowiki/internal/actors"
	"github.com/sidereusnuntius/gowiki/internal/federation"
)

func setupActivityPubHandler(actors *actors.Actors) *federation.FedGateway {
	return &federation.FedGateway{
		Actors: actors,
	}
}
