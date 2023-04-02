package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-session/session"
)

const (
	PathAuth           = "/auth"
	PathLogin          = "/login"
	PathOauthAuthorize = "/oauth/authorize"
	PathOauthToken     = "/oauth/token"
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
}

func (c *authController) Authorize(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/Authorize")
	defer span.End()

	store, err := session.Start(ctx, w, r.WithContext(ctx))
	if err != nil {
		httpError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	var form url.Values
	if v, ok := store.Get("ReturnUri"); ok {
		form = v.(url.Values)
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

	err := c.authProvider.HandleTokenRequest(w, r.WithContext(ctx))
	if err != nil {
		httpError(ctx, w, err, http.StatusBadRequest)
		return
	}
}

func (c *authController) UserAuthHandler(w http.ResponseWriter, r *http.Request) (string, error) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/UserAuthHandler")
	defer span.End()

	store, err := session.Start(ctx, w, r.WithContext(ctx))
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

		store.Set("ReturnUri", r.Form)
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

func (c *authController) PasswordAuthHandler(ctx context.Context, _, email, password string) (string, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/PasswordAuthHandler")
	defer span.End()

	/*user, err := c.userService.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if !pkg.IsPasswordMatching(user.Password, password) {
		return "", errors.ErrAuthCredentials
	}

	return user.Key, nil*/

	return "", nil
}

func (c *authController) ClientAuthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "transport.http.handler/AuthHandler")
	defer span.End()

	store, err := session.Start(nil, w, r)
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

	store, err := session.Start(r.Context(), w, r)
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

		key := ""
		/*if user, err := c.userService.GetByEmail(ctx, r.Form.Get("email")); err == nil {
			key = user.Key
		}*/

		store.Set("UserLoginKey", key)
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
			<input type="email" name="email" /> <small>try gabor@elemo.app</small><br>
			<input type="password" name="password" /> <small>try AppleTree123</small><br>
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
	}

	/*
		if controller.userService == nil {
			return nil, errors.ErrNoUserService
		}

		if controller.authProvider == nil {
			return nil, errors.ErrNoAuthProvider
		}
	*/

	controller.authProvider.SetUserAuthorizationHandler(controller.UserAuthHandler)
	controller.authProvider.SetPasswordAuthorizationHandler(controller.PasswordAuthHandler)

	return controller, nil
}
