package authhelpers

import (
	"context"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type sessionKey struct{}

func Authenticated(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r.Context()) {
			// TODO: encode the request's URL path and send it as a query parameter.
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// PutSession returns a new context containing the provided session.
func PutSession(ctx context.Context, session *model.Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, session)
}

func GetSession(ctx context.Context) (*model.Session, bool) {
	if session, ok := ctx.Value(sessionKey{}).(*model.Session); ok && session != nil {
		return session, true
	}
	return nil, false
}

func IsAuthenticated(ctx context.Context) bool {
	_, ok := GetSession(ctx)
	return ok
}
