package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	authErrors "github.com/go-oauth2/oauth2/v4/errors"
	authServer "github.com/go-oauth2/oauth2/v4/server"
	netHTTPMiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

const (
	PathRoot    = "/"
	PathSwagger = "/swagger.json"
	PathMetrics = "/metrics"

	DefaultRequestThrottleLimit   = 100
	DefaultRequestThrottleBacklog = 1000
	DefaultRequestThrottleTimeout = 10 * time.Second
)

// StrictServer is the type alias for the generated server interface.
type StrictServer interface {
	api.StrictServerInterface
	AuthController
	InternalErrorHandler(err error) (re *authErrors.Response)
	ResponseErrorHandler(r *authErrors.Response)
	PreRedirectErrorHandler(w http.ResponseWriter, r *authServer.AuthorizeRequest, err error)
}

// Server is the type alias for the generated server interface.
type Server interface {
	api.ServerInterface
	AuthController
	InternalErrorHandler(err error) *authErrors.Response
	ResponseErrorHandler(r *authErrors.Response)
	PreRedirectErrorHandler(w http.ResponseWriter, r *authServer.AuthorizeRequest, err error)
}

// server is the concrete implementation of the ServerInterface.
type server struct {
	*baseController

	AuthController
	OrganizationController
	NamespaceController
	ProjectController
	IssueController
	DocumentController
	FolderController
	LabelController
	UserController
	TodoController
	SystemController
	PermissionController
	NotificationController
	SearchController
}

func (s *server) InternalErrorHandler(err error) *authErrors.Response {
	return authErrors.NewResponse(err, http.StatusInternalServerError)
}

func (s *server) ResponseErrorHandler(r *authErrors.Response) {
	s.logger.Error(context.Background(), r.Description,
		log.WithError(r.Error),
		log.WithStatus(r.StatusCode),
		log.WithValue(r.ErrorCode),
	)
}

func (s *server) PreRedirectErrorHandler(_ http.ResponseWriter, r *authServer.AuthorizeRequest, err error) {
	s.logger.Error(context.Background(), err.Error(),
		log.WithError(err),
		log.WithUserID(r.UserID),
		log.WithAuthClientID(r.ClientID),
	)
}

// ServerDeps holds the required services and auth provider for NewServer.
type ServerDeps struct {
	AuthProvider        *authServer.Server
	OrganizationService service.OrganizationService
	NamespaceService    service.NamespaceService
	ProjectService      service.ProjectService
	IssueService        service.IssueService
	DocumentService     service.DocumentService
	FolderService       service.FolderService
	LabelService        service.LabelService
	RoleService         service.RoleService
	TeamService         service.TeamService
	UserService         service.UserService
	EmailService        service.EmailService
	TodoService         service.TodoService
	SystemService       service.SystemService
	LicenseService      service.LicenseService
	PermissionService   service.PermissionService
	NotificationService service.NotificationService
	SearchService       service.SearchService
}

// NewServer creates a new HTTP server.
func NewServer(deps ServerDeps, opts ...ControllerOption) (StrictServer, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	s := &server{
		baseController: c,
	}

	if s.AuthController, err = NewAuthController(deps.UserService, deps.AuthProvider, opts...); err != nil {
		return nil, err
	}

	if s.OrganizationController, err = NewOrganizationController(
		deps.OrganizationService,
		deps.RoleService,
		deps.TeamService,
		deps.UserService,
		opts...,
	); err != nil {
		return nil, err
	}

	if s.NamespaceController, err = NewNamespaceController(deps.NamespaceService, opts...); err != nil {
		return nil, err
	}

	if s.ProjectController, err = NewProjectController(deps.ProjectService, opts...); err != nil {
		return nil, err
	}

	if s.IssueController, err = NewIssueController(deps.IssueService, opts...); err != nil {
		return nil, err
	}

	if s.DocumentController, err = NewDocumentController(deps.DocumentService, opts...); err != nil {
		return nil, err
	}

	if s.FolderController, err = NewFolderController(deps.FolderService, opts...); err != nil {
		return nil, err
	}

	if s.LabelController, err = NewLabelController(deps.LabelService, opts...); err != nil {
		return nil, err
	}

	if s.UserController, err = NewUserController(deps.UserService, deps.EmailService, opts...); err != nil {
		return nil, err
	}

	if s.TodoController, err = NewTodoController(deps.TodoService, opts...); err != nil {
		return nil, err
	}

	if s.SystemController, err = NewSystemController(deps.SystemService, deps.LicenseService, opts...); err != nil {
		return nil, err
	}

	if s.PermissionController, err = NewPermissionController(deps.PermissionService, opts...); err != nil {
		return nil, err
	}

	if s.NotificationController, err = NewNotificationController(
		deps.NotificationService,
		deps.UserService,
		opts...,
	); err != nil {
		return nil, err
	}

	if s.SearchController, err = NewSearchController(deps.SearchService, opts...); err != nil {
		return nil, err
	}

	return s, nil
}

// NewRouter creates a new HTTP router for the Server.
func NewRouter(strictServer StrictServer, serverConfig *config.ServerConfig, tracer tracing.Tracer) (http.Handler, error) {
	if serverConfig == nil {
		return nil, config.ErrNoConfig
	}

	swagger, err := api.GetSpec()
	if err != nil {
		return nil, errors.Join(ErrInvalidSwagger, err)
	}

	swagger.Servers = nil

	throttleLimit := DefaultRequestThrottleLimit
	if serverConfig.RequestThrottleLimit > 0 {
		throttleLimit = serverConfig.RequestThrottleLimit
	}

	throttleBacklog := DefaultRequestThrottleBacklog
	if serverConfig.RequestThrottleBacklog > 0 {
		throttleBacklog = serverConfig.RequestThrottleBacklog
	}

	throttleTimeout := DefaultRequestThrottleTimeout
	if serverConfig.RequestThrottleTimeout > 0 {
		throttleTimeout = serverConfig.RequestThrottleTimeout * time.Second
	}

	s := api.NewStrictHandler(strictServer, nil)

	router := chi.NewRouter()

	router.Use(
		WithPrometheusMetrics,
		WithOtelTracer,
		middleware.ThrottleBacklog(throttleLimit, throttleBacklog, throttleTimeout),
		middleware.RequestID,
		// Single trusted reverse-proxy hop: use the rightmost X-Forwarded-For
		// entry. Falls back to RemoteAddr in request logging when unset.
		middleware.ClientIPFromXFF(),
		middleware.AllowContentEncoding("deflate", "gzip"),
		middleware.Compress(5, "text/html", "text/css", "application/json"),
		middleware.SetHeader("X-Frame-Options", "sameorigin"),
		middleware.StripSlashes,
		WithTracedMiddleware(tracer, WithRequestLogger),
		middleware.Recoverer,
	)

	if serverConfig.CORS.Enabled {
		router.Use(cors.Handler(cors.Options{
			AllowedOrigins:   serverConfig.CORS.AllowedOrigins,
			AllowedMethods:   serverConfig.CORS.AllowedMethods,
			AllowedHeaders:   serverConfig.CORS.AllowedHeaders,
			AllowCredentials: serverConfig.CORS.AllowCredentials,
			MaxAge:           serverConfig.CORS.MaxAge,
		}))
	}

	router.Group(func(r chi.Router) {
		r.Use(
			WithTracedMiddleware(tracer, WithUserID(strictServer.ValidateBearerToken)),
			WithTracedMiddleware(tracer, netHTTPMiddleware.OapiRequestValidatorWithOptions(swagger, &netHTTPMiddleware.Options{
				Options: openapi3filter.Options{
					AuthenticationFunc: func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
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

		r.Handle(PathRoot, api.HandlerFromMux(s, r))
	})

	router.Handle(PathAuth, http.HandlerFunc(strictServer.ClientAuthHandler))
	router.Handle(PathLogin, http.HandlerFunc(strictServer.LoginHandler))
	router.Handle(PathOauthAuthorize, http.HandlerFunc(strictServer.Authorize))
	router.Handle(PathOauthToken, http.HandlerFunc(strictServer.Token))

	router.Handle(PathSwagger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "transport.http.handler/GetSwagger")
		defer span.End()

		WriteJSONResponse(w, swagger, http.StatusOK)
	}))

	return router, nil
}

// NewMetricsServer creates a new HTTP server for Prometheus metrics.
func NewMetricsServer(serverConfig *config.ServerConfig, tracer tracing.Tracer) (http.Handler, error) {
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

	router.Route(PathMetrics, func(r chi.Router) {
		r.Handle(PathRoot, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "transport.http.handler/GetPrometheusMetrics")
			defer span.End()

			promhttp.Handler().ServeHTTP(w, r.WithContext(ctx))
		}))
	})

	return router, nil
}
