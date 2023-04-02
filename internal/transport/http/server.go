package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-oauth2/oauth2/v4"
	authErrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/transport/http/gen"
)

var (
	ErrNoConfig         = errors.New("no config provided")       // no config provided
	ErrInvalidSwagger   = errors.New("invalid swagger provided") // invalid swagger provided
	ErrAuthNoPermission = errors.New("no permission")            // no permission
	ErrAuthCredentials  = errors.New("invalid credentials")      // invalid credentials
)

// StrictServer is the type alias for the generated server interface.
type StrictServer interface {
	gen.StrictServerInterface
	AuthController
	InternalErrorHandler(err error) (re *authErrors.Response)
	ResponseErrorHandler(re *authErrors.Response)
}

// Server is the type alias for the generated server interface.
type Server interface {
	gen.ServerInterface
	AuthController
	InternalErrorHandler(err error) *authErrors.Response
	ResponseErrorHandler(r *authErrors.Response)
}

// server is the concrete implementation of the ServerInterface.
type server struct {
	*baseController

	authController   AuthController
	systemController SystemController
}

func (s *server) Authorize(w http.ResponseWriter, r *http.Request) {
	s.authController.Authorize(w, r)
}

func (s *server) Token(w http.ResponseWriter, r *http.Request) {
	s.authController.Token(w, r)
}

func (s *server) PasswordAuthHandler(ctx context.Context, clientID, email, password string) (string, error) {
	return s.authController.PasswordAuthHandler(ctx, clientID, email, password)
}

func (s *server) UserAuthHandler(w http.ResponseWriter, r *http.Request) (string, error) {
	return s.authController.UserAuthHandler(w, r)
}

func (s *server) ClientAuthHandler(w http.ResponseWriter, r *http.Request) {
	s.authController.ClientAuthHandler(w, r)
}
func (s *server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	s.authController.LoginHandler(w, r)
}

func (s *server) ValidateBearerToken(r *http.Request) (oauth2.TokenInfo, error) {
	return s.authController.ValidateBearerToken(r)
}

func (s *server) ValidateTokenHandler(r *http.Request) error {
	return s.authController.ValidateTokenHandler(r)
}

func (s *server) GetSystemHealth(ctx context.Context, request gen.GetSystemHealthRequestObject) (gen.GetSystemHealthResponseObject, error) {
	return s.systemController.GetSystemHealth(ctx, request)
}

func (s *server) GetSystemHeartbeat(ctx context.Context, request gen.GetSystemHeartbeatRequestObject) (gen.GetSystemHeartbeatResponseObject, error) {
	return s.systemController.GetSystemHeartbeat(ctx, request)
}

func (s *server) GetSystemVersion(ctx context.Context, request gen.GetSystemVersionRequestObject) (gen.GetSystemVersionResponseObject, error) {
	return s.systemController.GetSystemVersion(ctx, request)
}

func (s *server) InternalErrorHandler(err error) *authErrors.Response {
	return authErrors.NewResponse(err, http.StatusInternalServerError)
}

func (s *server) ResponseErrorHandler(r *authErrors.Response) {
	s.logger.Error(r.Description,
		log.WithError(r.Error),
		log.WithStatus(r.StatusCode),
		log.WithValue(r.ErrorCode),
	)
}

// NewServer creates a new HTTP server.
func NewServer(opts ...ControllerOption) (StrictServer, error) {
	var err error

	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	s := &server{
		baseController: c,
	}

	if s.authController, err = NewAuthController(opts...); err != nil {
		return nil, err
	}

	if s.systemController, err = NewSystemController(opts...); err != nil {
		return nil, err
	}

	return s, nil
}

// NewRouter creates a new HTTP router for the Server.
func NewRouter(strictServer StrictServer, serverConfig *config.ServerConfig, tracer trace.Tracer) (http.Handler, error) {
	if serverConfig == nil {
		return nil, ErrNoConfig
	}

	swagger, err := gen.GetSwagger()
	if err != nil {
		return nil, errors.Join(ErrInvalidSwagger, err)
	}

	swagger.Servers = openapi3.Servers{
		&openapi3.Server{
			URL:         fmt.Sprintf("https://%s", serverConfig.Address),
			Description: "Default server",
		},
		&openapi3.Server{
			URL:         "{url}",
			Description: "Third-party server",
			Variables: map[string]*openapi3.ServerVariable{
				"url": {
					Default: "https://example.com/api",
				},
			},
		},
	}

	s := gen.NewStrictHandler(strictServer, nil)

	router := chi.NewRouter()

	router.Use(
		WithPrometheusMetrics,
		WithOtelTracer,
		WithTracedMiddleware(tracer, middleware.RequestID),
		WithTracedMiddleware(tracer, middleware.RealIP),
		WithTracedMiddleware(tracer, middleware.AllowContentEncoding("deflate", "gzip")),
		WithTracedMiddleware(tracer, middleware.Compress(7, "text/html", "text/css", "application/json")),
		WithTracedMiddleware(tracer, middleware.SetHeader("X-Frame-Options", "sameorigin")),
		WithTracedMiddleware(tracer, WithRequestLogger),
		WithTracedMiddleware(tracer, middleware.Recoverer),
	)

	if serverConfig.CORS.Enabled {
		router.Use(WithTracedMiddleware(tracer, cors.Handler(cors.Options{
			AllowedOrigins:   serverConfig.CORS.AllowedOrigins,
			AllowedMethods:   serverConfig.CORS.AllowedMethods,
			AllowedHeaders:   serverConfig.CORS.AllowedHeaders,
			AllowCredentials: serverConfig.CORS.AllowCredentials,
			MaxAge:           serverConfig.CORS.MaxAge,
		})))
	}

	router.Group(func(r chi.Router) {
		r.Use(
			WithTracedMiddleware(tracer, WithUserKey(strictServer.ValidateBearerToken)),
			WithTracedMiddleware(tracer, oapiMiddleware.OapiRequestValidatorWithOptions(swagger, &oapiMiddleware.Options{
				Options: openapi3filter.Options{
					AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
						if err := strictServer.ValidateTokenHandler(input.RequestValidationInput.Request); err != nil {
							if errors.Is(err, authErrors.ErrInvalidAccessToken) {
								return ErrAuthNoPermission
							}

							return ErrAuthCredentials
						}

						return nil
					},
				},
			})),
		)

		r.Handle("/", gen.HandlerFromMux(s, r))
	})

	router.Handle(PathAuth, http.HandlerFunc(strictServer.ClientAuthHandler))
	router.Handle(PathLogin, http.HandlerFunc(strictServer.LoginHandler))
	router.Handle(PathOauthAuthorize, http.HandlerFunc(strictServer.Authorize))
	router.Handle(PathOauthToken, http.HandlerFunc(strictServer.Token))

	router.Handle("/swagger.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "transport.http.handler/GetSwagger")
		defer span.End()

		WriteJSONResponse(w, swagger, http.StatusOK)
	}))

	return router, nil
}

// NewMetricsServer creates a new HTTP server for Prometheus metrics.
func NewMetricsServer(serverConfig *config.ServerConfig, tracer trace.Tracer) (http.Handler, error) {
	router := chi.NewRouter()

	if serverConfig.CORS.Enabled {
		router.Use(WithTracedMiddleware(tracer, cors.Handler(cors.Options{
			AllowedOrigins:   serverConfig.CORS.AllowedOrigins,
			AllowedMethods:   serverConfig.CORS.AllowedMethods,
			AllowedHeaders:   serverConfig.CORS.AllowedHeaders,
			AllowCredentials: serverConfig.CORS.AllowCredentials,
			MaxAge:           serverConfig.CORS.MaxAge,
		})))
	}

	router.Route("/metrics", func(r chi.Router) {
		r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "transport.http.handler/GetPrometheusMetrics")
			defer span.End()

			promhttp.Handler().ServeHTTP(w, r.WithContext(ctx))
		}))
	})

	return router, nil
}
