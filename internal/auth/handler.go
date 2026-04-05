package auth

import (
	"net/http"
	"time"

	authhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/auth"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/render"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

const cookieName = "sessionId"

type Handler struct {
	AuthService  *Auth
	SessionStore SessionStore
}

func (handler *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /register", handler.RegisterAction)
	mux.HandleFunc("GET /register", handler.Register)
	mux.HandleFunc("GET /login", handler.Login)
	mux.HandleFunc("POST /login", handler.LoginAction)
	mux.HandleFunc("POST /logout", authhelpers.Authenticated(handler.Logout))
}

func (handler *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to initialize request data")
		return
	}

	err = handler.SessionStore.DeleteSession(p.Ctx, p.Content.Session.Token)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to delete session after logout")
	}

	ClearToken(w)
	p.Content.Session = nil
	p.Content.Authenticated = false
	// TODO: redirect the user to the page in which they clicked logout.
	p.ReloadPage("auth/login.html")
}

func (handler *Handler) Login(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to initialize request data")
	}
	p.Content.Title = "Login"

	p.Render("auth/login.html")
}

func (handler *Handler) LoginAction(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to initialize request data")
		return
	}

	in := model.LoginInput{
		Email:    p.GetString("email"),
		Password: p.GetString("password"),
	}

	session, err := handler.AuthService.Authenticate(p.Ctx, in)
	if err != nil {
		p.HandleError(err)
		return
	}

	SetToken(w, &session)
	p.Content.Session = &session
	p.Content.Authenticated = true
	p.ReloadPage("main.html")
	// TODO: when the user successfully logs in or registers, we want to redirect them to the page they were before, but
	// we will set a query parameter which will triger the full reload of the #page-container.
}

func SetToken(w http.ResponseWriter, session *model.Session) {
	cookie := http.Cookie{
		Name:    cookieName,
		Value:   session.Token,
		Expires: session.Expiration,
		MaxAge:  int(time.Until(session.Expiration).Seconds()),
	}

	http.SetCookie(w, &cookie)
}

func ClearToken(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:   cookieName,
		Value:  "",
		MaxAge: -1,
	}

	http.SetCookie(w, &cookie)
}

func (handler *Handler) RegisterAction(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to initialize request data")
	}

	in := model.RegisterInput{
		Username: p.GetString("username"),
		Email:    p.GetString("email"),
		Password: p.GetString("password"),
	}

	session, err := handler.AuthService.RegisterUser(r.Context(), in, false)
	if err != nil {
		p.HandleError(err)
		return
	}

	if len(session.Token) > 0 {
		SetToken(w, &session)
		p.Content.Session = &session
		p.Content.Authenticated = true
	}

	p.ReloadPage("main.html")
}

func (handler *Handler) Register(w http.ResponseWriter, r *http.Request) {
	page, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to initialize request data")
		return
	}

	page.Content.Title = "Register"
	page.Render("auth/register.html")
}
