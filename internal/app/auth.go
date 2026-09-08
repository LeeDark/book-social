package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	httpauth "github.com/LeeDark/book-social/internal/http/auth"
	"github.com/LeeDark/book-social/internal/http/flash"
	"github.com/LeeDark/book-social/internal/http/render"
	"github.com/LeeDark/book-social/internal/http/response"
	"github.com/LeeDark/book-social/internal/http/view"
	"github.com/LeeDark/book-social/internal/modules/users"
)

type authUserService interface {
	RegisterAndCreateSession(context.Context, users.RegistrationInput, []byte, time.Duration) (users.User, error)
	Authenticate(context.Context, string, string) (users.User, error)
}

type sessionService interface {
	CreateSession(context.Context, int, []byte) error
	DeleteSession(context.Context, []byte) error
}

type AuthHandler struct {
	users    authUserService
	sessions sessionService
	cookies  *httpauth.CookieManager
	flashes  *flash.Manager
	renderer *render.Renderer
	logger   *slog.Logger
	lifetime time.Duration
}

type AuthPageData struct {
	view.Page
	Form AuthForm
}

type AuthForm struct {
	FirstName  string
	Login      string
	Email      string
	Identifier string
	Errors     map[string]string
}

func NewAuthHandler(userService authUserService, sessionService sessionService, cookies *httpauth.CookieManager, flashes *flash.Manager, renderer *render.Renderer, logger *slog.Logger, lifetime time.Duration) *AuthHandler {
	return &AuthHandler{users: userService, sessions: sessionService, cookies: cookies, flashes: flashes, renderer: renderer, logger: logger, lifetime: lifetime}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if h.redirectAuthenticated(w, r) {
		return
	}
	if r.Method == http.MethodGet {
		h.renderRegister(w, r, http.StatusOK, AuthForm{})
		return
	}
	if r.Method != http.MethodPost {
		response.ClientError(w, http.StatusMethodNotAllowed)
		return
	}

	input := users.RegistrationInput{FirstName: r.FormValue("first_name"), Login: r.FormValue("login"), Email: r.FormValue("email"), Password: r.FormValue("password"), PasswordConfirmation: r.FormValue("password_confirmation")}
	form := AuthForm{FirstName: input.FirstName, Login: input.Login, Email: input.Email}
	token, err := h.cookies.GenerateToken()
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("generate registration session token: %w", err))
		return
	}
	if _, err = h.users.RegisterAndCreateSession(r.Context(), input, httpauth.HashToken(token), h.lifetime); err != nil {
		if h.renderRegistrationError(w, r, form, err) {
			return
		}
		response.ServerError(w, r, h.logger, fmt.Errorf("register account: %w", err))
		return
	}
	if err := h.cookies.Set(w, token); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("set registration session cookie: %w", err))
		return
	}
	h.flashes.Set(w, flash.Registered)
	http.Redirect(w, r, "/me", http.StatusSeeOther)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.redirectAuthenticated(w, r) {
		return
	}
	if r.Method == http.MethodGet {
		h.renderLogin(w, r, http.StatusOK, AuthForm{})
		return
	}
	if r.Method != http.MethodPost {
		response.ClientError(w, http.StatusMethodNotAllowed)
		return
	}

	form := AuthForm{Identifier: r.FormValue("identifier")}
	user, err := h.users.Authenticate(r.Context(), form.Identifier, r.FormValue("password"))
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredentials) {
			form.Errors = map[string]string{"credentials": "Invalid login or password."}
			h.renderLogin(w, r, http.StatusUnprocessableEntity, form)
			return
		}
		response.ServerError(w, r, h.logger, fmt.Errorf("authenticate account: %w", err))
		return
	}
	token, err := h.cookies.GenerateToken()
	if err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("generate login session token: %w", err))
		return
	}
	if err = h.sessions.CreateSession(r.Context(), user.ID, httpauth.HashToken(token)); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("create login session: %w", err))
		return
	}
	if err = h.cookies.Set(w, token); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("set login session cookie: %w", err))
		return
	}
	h.flashes.Set(w, flash.SignedIn)
	http.Redirect(w, r, "/me", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.ClientError(w, http.StatusMethodNotAllowed)
		return
	}
	if token, ok := h.cookies.ReadToken(r); ok {
		if err := h.sessions.DeleteSession(r.Context(), httpauth.HashToken(token)); err != nil && !errors.Is(err, users.ErrUnauthenticated) {
			response.ServerError(w, r, h.logger, fmt.Errorf("delete login session: %w", err))
			return
		}
	}
	h.cookies.Clear(w)
	h.flashes.Set(w, flash.SignedOut)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	identity, _ := httpauth.CurrentUserFromRequest(r)
	data := AuthPageData{Page: view.Page{Title: "Your account", Nav: view.MainNavigation()}, Form: AuthForm{}}
	view.ApplyRequestState(&data.Page, r)
	data.Page.CurrentUser = &view.CurrentUser{ID: identity.ID, DisplayName: identity.FirstName}
	if err := h.renderer.Render(w, http.StatusOK, "me.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render account page: %w", err))
	}
}

func (h *AuthHandler) redirectAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := httpauth.CurrentUserFromRequest(r); !ok {
		return false
	}
	http.Redirect(w, r, "/me", http.StatusSeeOther)
	return true
}

func (h *AuthHandler) renderRegister(w http.ResponseWriter, r *http.Request, status int, form AuthForm) {
	data := AuthPageData{Page: view.Page{Title: "Register", Nav: view.MainNavigation()}, Form: form}
	view.ApplyRequestState(&data.Page, r)
	if err := h.renderer.Render(w, status, "register.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render registration page: %w", err))
	}
}

func (h *AuthHandler) renderLogin(w http.ResponseWriter, r *http.Request, status int, form AuthForm) {
	data := AuthPageData{Page: view.Page{Title: "Login", Nav: view.MainNavigation()}, Form: form}
	view.ApplyRequestState(&data.Page, r)
	if err := h.renderer.Render(w, status, "login.tmpl", data); err != nil {
		response.ServerError(w, r, h.logger, fmt.Errorf("render login page: %w", err))
	}
}

func (h *AuthHandler) renderRegistrationError(w http.ResponseWriter, r *http.Request, form AuthForm, err error) bool {
	if validation, ok := errors.AsType[users.ValidationError](err); ok {
		form.Errors = map[string]string{validation.Field: validation.Message}
		h.renderRegister(w, r, http.StatusUnprocessableEntity, form)
		return true
	}
	if errors.Is(err, users.ErrLoginTaken) {
		form.Errors = map[string]string{"login": "is already in use"}
		h.renderRegister(w, r, http.StatusUnprocessableEntity, form)
		return true
	}
	if errors.Is(err, users.ErrEmailTaken) {
		form.Errors = map[string]string{"email": "is already in use"}
		h.renderRegister(w, r, http.StatusUnprocessableEntity, form)
		return true
	}
	return false
}
