package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/config"
	authhelpers "github.com/sidereusnuntius/gowiki/internal/helpers/auth"
	httphelpers "github.com/sidereusnuntius/gowiki/internal/helpers/http"
	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/render"
	"github.com/sidereusnuntius/gowiki/internal/view"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

const cookieName = "sessionId"

type Handler struct {
	Config       *config.WikiConfig
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
	p.Redirect("/", "#content")
	p.PatchElement("authheader.html", "auth-header", "#auth-header")
	// p.ReloadPage("auth/login.html")
}

func (handler *Handler) Login(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to initialize request data")
	}

	var view view.AuthView

	proceed, ok := parseProceedArg(r.URL.Query().Get("proceed"))
	if ok {
		view.SuccessRedirect = proceed
	}

	referer, ok := httphelpers.LocalReferer(r, handler.Config.URL.Host)
	if ok {
		view.SuccessRedirect = referer
	}

	p.Content.Data = view
	p.Content.Title = "Login"

	p.Render("auth/login.html")
}

func parseProceedArg(encoded string) (string, bool) {
	if len(encoded) == 0 {
		return "", false
	}

	proceedBytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		wikilog.Logger.Error().
			Err(err).
			Str("proceed", encoded).
			Msg("error parsing login redirect URL")
		return "", false
	}
	proceed := string(proceedBytes)

	if !strings.HasPrefix(proceed, "/") && strings.Contains(proceed, "//") {
		return "", false
	}

	return proceed, true
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

	p.Content.Session = &session
	p.Content.Authenticated = true
	SetToken(w, &session)

	proceed := p.GetString("proceed")
	if len(proceed) == 0 {
		proceed = "/"
	}
	p.Redirect(proceed, "#content")
	p.PatchElement("authheader.html", "auth-header", "#auth-header")

	// p.ReloadPage("main.html")
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

	p.Content.Session = &session
	p.Content.Authenticated = true
	SetToken(w, &session)

	proceed := p.GetString("proceed")
	if len(proceed) == 0 {
		proceed = "/"
	}
	p.Redirect(proceed, "#content")
	p.PatchElement("authheader.html", "auth-header", "#auth-header")
}

func (handler *Handler) Register(w http.ResponseWriter, r *http.Request) {
	page, err := render.Init(w, r)
	if err != nil {
		wikilog.Logger.Error().Err(err).Msg("failed to initialize request data")
		return
	}

	var view view.AuthView

	referer, ok := httphelpers.LocalReferer(r, handler.Config.URL.Host)
	if ok && referer != "/login" {
		view.SuccessRedirect = referer
	}

	page.Content.Data = view
	page.Content.Title = "Register"
	page.Render("auth/register.html")
}
