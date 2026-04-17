package setup

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/actors"
	"github.com/sidereusnuntius/gowiki/internal/articles"
	"github.com/sidereusnuntius/gowiki/internal/federation"
	hostsql "github.com/sidereusnuntius/gowiki/internal/federation/store"
	"github.com/sidereusnuntius/gowiki/internal/keystore"
)

func setupHostsStore(db *sql.DB) *hostsql.HostsStore {
	return hostsql.New(db)
}

func setupActivityPubHandler(fed *federation.Federation, articles *articles.ArticleService, actors *actors.Actors, keys *keystore.KeyStore) *federation.FedGateway {
	return &federation.FedGateway{
		Actors:     actors,
		Articles:   articles,
		Keys:       keys,
		Federation: fed,
	}
}
