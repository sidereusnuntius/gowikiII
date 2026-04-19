package federation

import (
	"io"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/federation/streams"
	httphelpers "github.com/sidereusnuntius/gowiki/internal/helpers/http"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/model/activitystreams"
	"github.com/sidereusnuntius/gowiki/internal/sanitize"
	"github.com/sidereusnuntius/gowiki/internal/security"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type FedGateway struct {
	Actors     ActorService
	Keys       *security.Security
	Articles   ArticleService
	Federation *Federation
}

func NewGateway(federation *Federation, articles ArticleService, actors ActorService, keys *security.Security) *FedGateway {
	return &FedGateway{
		Actors:     actors,
		Keys:       keys,
		Articles:   articles,
		Federation: federation,
	}
}

func (fg *FedGateway) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /u/{username}", fg.GetActor)
	mux.HandleFunc("GET /u/{username}/main-key", fg.GetActor) // TODO: return only basic information about the actor.
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

		next.ServeHTTP(w, r)
	}
}

func (fg *FedGateway) Inbox(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// TODO: better handle errors.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		wikilog.Logger.Error().Err(err).Msg("failed to read activity received at inbox")
		return
	}

	activity, err := activitystreams.ReadActivity(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		wikilog.Logger.Error().Err(err).Msg("failed to parse activity received at inbox")
		return
	}

	if err = fg.Federation.ProcessActivity(r.Context(), activity); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		wikilog.Logger.Error().Err(err).Msgf("failed to process %s activity", activity.Type)
		return
	}
}
