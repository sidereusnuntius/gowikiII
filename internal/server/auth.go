package server

import (
	"fmt"
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/model"
	"github.com/sidereusnuntius/gowiki/internal/render"
	"github.com/sidereusnuntius/gowiki/internal/wiki"
)

type AuthHandler struct {
	AuthService *wiki.Auth
}

func (handler *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /register", handler.RegisterAction)
	mux.HandleFunc("GET /register", handler.Register)
}

func (handler *AuthHandler) RegisterAction(w http.ResponseWriter, r *http.Request) {
	p, err := render.Init(w, r)
	if err != nil {
		fmt.Println("failed to read form body:", err)
	}

	in := model.RegisterInput{
		Username: p.GetString("username"),
		Email:    p.GetString("email"),
		Password: p.GetString("password"),
	}

	err = handler.AuthService.RegisterUser(r.Context(), in, false)
	if err != nil {
		p.RenderText("signup-result", "Error: " + err.Error())
		return
	}
	p.RenderText("signup-result", "Successfully registered!")
}

func (handler *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	page, err := render.Init(w, r)
	if err != nil {
		fmt.Println("failed to initialize page data:", err)
	}

	page.Title = "Register"
	if err := page.Render("auth/register.html"); err != nil {
		fmt.Println(err)
	}
}
