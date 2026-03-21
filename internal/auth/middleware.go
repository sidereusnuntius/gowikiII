package auth

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	authhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/auth"
)

func (handler *Handler) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil || cookie == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		token := cookie.Value

		if len(token) > 0 {
			session, err := handler.SessionStore.GetFullSession(ctx, token)
			switch {
			case err != nil:
				ClearToken(w) // TODO: only clear if token is not found.
				log.Error().Err(err).Msg("failed to retrieve session from session store")
			case session.Expiration.Before(time.Now()):
				log.Debug().Msg("cleaning up expired sessions")
				ClearToken(w)
				if err = handler.SessionStore.DeleteSession(ctx, token); err != nil {
					log.Error().Err(err).Str("token", token).Msg("failed to delete session from session store")
				}
			default:
				// TODO: define a save method that allows handlers to update the session and restart the TTL. This can be useful if I later want to store additional data in the session.
				ctx = authhelpers.PutSession(ctx, &session)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}
