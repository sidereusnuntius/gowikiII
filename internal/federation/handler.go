package federation

import (
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/federation/streams"
	httphelpers "github.com/sidereusnuntius/gowiki/internal/helpers/http"
	"github.com/sidereusnuntius/gowiki/internal/keystore"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/sanitize"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type FedGateway struct {
	Actors     ActorService
	Keys       *keystore.KeyStore
	Articles   ArticleService
	Federation *Federation
}

func NewGateway(federation *Federation, articles ArticleService, actors ActorService, keys *keystore.KeyStore) *FedGateway {
	return &FedGateway{
		Actors:     actors,
		Keys:       keys,
		Articles:   articles,
		Federation: federation,
	}
}

func (fg *FedGateway) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /u/{username}", fg.GetActor)
	mux.HandleFunc("GET /a/{slug}", fg.GetArticle)
	mux.HandleFunc("POST /inbox", fg.SignedMiddleware(fg.Inbox))
}

func (fg *FedGateway) GetArticle(w http.ResponseWriter, r *http.Request) {
	slug := sanitize.Normalize(
		r.PathValue("slug"),
	)

	req := model.ArticleRequest{
		Slug: slug,
	}
	article, err := fg.Articles.ArticleContent(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // TODO: handle error
		return
	}

	articleAS := streams.ArticleAS(&article)
	if err = httphelpers.WriteActivity(w, articleAS); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		wikilog.Logger.Error().Err(err).Msg("httphelpers.WriteActivity(w, articleAS)")
	}
}

func (fg *FedGateway) GetActor(w http.ResponseWriter, r *http.Request) {
	username := sanitize.Normalize(
		r.PathValue("username"),
	)

	actor, err := fg.Actors.GetLocalActor(r.Context(), username)
	if err != nil {
		// do something
		wikilog.Logger.Error().Err(err).Msg("fg.Actors.GetLocalActor()")
		return
	}

	actorAS := streams.ActorAS(&actor)

	if err = httphelpers.WriteActivity(w, actorAS); err != nil {
		wikilog.Logger.Error().Err(err).Msg("GetActor()")
	}
}

func (fg *FedGateway) SignedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fg.Keys.VerifySignature(r.Context(), r); err != nil {
			wikilog.Logger.Error().Err(err).Msg("failed to verify signature for request")
			w.Write([]byte("invalid signature")) // TODO: improve error handling
		}

	}
}

func (fg *FedGateway) Inbox(w http.ResponseWriter, r *http.Request) {

}
