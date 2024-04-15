package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-session/session"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/password"
)

const (
	PathAuth           = "/auth"
	PathLogin          = "/login"
	PathOauthAuthorize = "/oauth/authorize"
	PathOauthToken     = "/oauth/token" // #nosec
)

// AuthController provides handlers for authentication and authorization.
type AuthController interface {
	// Authorize handles the authorization of a request.
	Authorize(w http.ResponseWriter, r *http.Request)
	// Token handles the generation and renewal of a token.
	Token(w http.ResponseWriter, r *http.Request)
	// PasswordAuthHandler handles the authorization of a user with a password.
	// If the authorization is successful, the user ID is returned. Otherwise,
	// an error is returned instead.
	PasswordAuthHandler(ctx context.Context, clientID, email, password string) (string, error)
	// UserAuthHandler handles the authorization of a user. If the
	// authorization is successful, the user ID is returned. Otherwise, an
	// error is returned instead.
	UserAuthHandler(w http.ResponseWriter, r *http.Request) (string, error)
	// ClientAuthHandler handles the client authorization.
	ClientAuthHandler(w http.ResponseWriter, r *http.Request)
	// LoginHandler handles the login of a user.
	LoginHandler(w http.ResponseWriter, r *http.Request)
	// ValidateBearerToken validates a bearer token and returns the token info
	// if the token is valid. Otherwise, an error is returned instead.
	ValidateBearerToken(r *http.Request) (oauth2.TokenInfo, error)
	// ValidateTokenHandler handles the validation of a token.
	ValidateTokenHandler(r *http.Request) error
}

type authController struct {
	*baseController
	sessionManager *session.Manager
}

func (c *authController) Authorize(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/Authorize")
	defer span.End()

	store, err := c.sessionManager.Start(ctx, w, r.WithContext(ctx))
	if err != nil {
		httpError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	var form url.Values
	if v, ok := store.Get("ReturnUri"); ok {
		if query, err := url.ParseQuery(v.(string)); err == nil {
			form = query
		}
	}

	r.Form = form

	store.Delete("ReturnUri")
	if err := store.Save(); err != nil {
		httpError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	err = c.authProvider.HandleAuthorizeRequest(w, r.WithContext(ctx))
	if err != nil {
		httpError(ctx, w, err, http.StatusBadRequest)
	}
}

func (c *authController) Token(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/Token")
	defer span.End()

	if err := c.authProvider.HandleTokenRequest(w, r.WithContext(ctx)); err != nil {
		httpError(ctx, w, err, http.StatusBadRequest)
		return
	}
}

func (c *authController) UserAuthHandler(w http.ResponseWriter, r *http.Request) (string, error) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/UserAuthHandler")
	defer span.End()

	store, err := c.sessionManager.Start(ctx, w, r.WithContext(ctx))
	if err != nil {
		return "", err
	}

	uid, ok := store.Get("UserLoginKey")

	if !ok {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				return "", err
			}
		}

		store.Set("ReturnUri", r.Form.Encode())
		if err := store.Save(); err != nil {
			return "", err
		}

		w.Header().Set("Location", PathLogin)
		w.WriteHeader(http.StatusFound)
		return "", nil
	}

	store.Delete("UserLoginKey")
	if err := store.Save(); err != nil {
		return "", err
	}

	return uid.(string), nil
}

func (c *authController) PasswordAuthHandler(ctx context.Context, _, email, passwd string) (string, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/PasswordAuthHandler")
	defer span.End()

	user, err := c.userService.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if !password.IsPasswordMatching(user.Password, passwd) || user.Status != model.UserStatusActive {
		return "", ErrAuthCredentials
	}

	return user.ID.String(), nil
}

func (c *authController) ClientAuthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/AuthHandler")
	defer span.End()

	store, err := c.sessionManager.Start(ctx, w, r)
	if err != nil {
		httpError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	if _, ok := store.Get("UserLoginKey"); !ok {
		w.Header().Set("Location", PathLogin)
		w.WriteHeader(http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	_, _ = w.Write([]byte(`<h1>Authorize page</h1>`))
	_, _ = w.Write([]byte(fmt.Sprintf(`
		<form action="%s" method="post">
			<h1>Authorize</h1>
			<p>The client would like to perform actions on your behalf.</p>

			<button type="submit" style="width:200px;">
				Allow
			</button>
		</form>
	`, PathOauthAuthorize)))
}

func (c *authController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/LoginHandler")
	defer span.End()

	store, err := c.sessionManager.Start(r.Context(), w, r)
	if err != nil {
		httpError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				httpError(ctx, w, err, http.StatusInternalServerError)
				return
			}
		}

		id := ""
		if user, err := c.userService.GetByEmail(ctx, r.Form.Get("email")); err == nil {
			id = user.ID.String()
		}

		store.Set("UserLoginKey", id)
		if err := store.Save(); err != nil {
			httpError(ctx, w, err, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", PathAuth)
		w.WriteHeader(http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	_, _ = w.Write([]byte(`<h1>Login page</h1>`))
	_, _ = w.Write([]byte(`
		<form method="post">
			<input type="email" name="email" /> <small></small><br>
			<input type="password" name="password" /> <small></small><br>
			<input type="submit" value="login">
		</form>
	`))
}

func (c *authController) ValidateBearerToken(r *http.Request) (oauth2.TokenInfo, error) {
	_, span := c.tracer.Start(r.Context(), "transport.http.handler/ValidateBearerToken")
	defer span.End()

	return c.authProvider.ValidationBearerToken(r)
}

func (c *authController) ValidateTokenHandler(r *http.Request) error {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/ValidateTokenHandler")
	defer span.End()

	_, err := c.ValidateBearerToken(r.WithContext(ctx))
	return err
}

// NewAuthController creates a new AuthController.
func NewAuthController(opts ...ControllerOption) (AuthController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &authController{
		baseController: c,
		sessionManager: session.NewManager(
			session.SetCookieName(c.conf.Session.CookieName),
			session.SetCookieLifeTime(c.conf.Session.MaxAge),
			session.SetSecure(c.conf.Session.Secure),
			session.SetEnableSIDInHTTPHeader(true),
		),
	}

	if controller.userService == nil {
		return nil, ErrNoUserService
	}

	if controller.authProvider == nil {
		return nil, ErrNoAuthProvider
	}

	controller.authProvider.SetUserAuthorizationHandler(controller.UserAuthHandler)
	controller.authProvider.SetPasswordAuthorizationHandler(controller.PasswordAuthHandler)

	return controller, nil
}
