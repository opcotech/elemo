package service

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
)

// Test-only aliases for unexported internals used by package service_test.

type Runtime = runtime

func NewRuntimeForTest(logger log.Logger, tracer tracing.Tracer) Runtime {
	return Runtime{
		logger: logger,
		tracer: tracer,
	}
}

func RuntimeLogger(r Runtime) log.Logger {
	return r.logger
}

func RuntimeTracer(r Runtime) tracing.Tracer {
	return r.tracer
}

func SystemServiceVersion(s SystemService) *model.VersionInfo {
	return s.(*systemService).versionInfo
}

func SetSystemServiceState(
	s SystemService,
	versionInfo *model.VersionInfo,
	resources map[model.HealthCheckComponent]Pingable,
) {
	svc := s.(*systemService)
	svc.versionInfo = versionInfo
	svc.resources = resources
}

func SetNotificationServiceRepo(s NotificationService, repo repository.NotificationRepository) {
	s.(*notificationService).notificationRepo = repo
}

func SetStaticFileServiceRepo(s StaticFileService, repo repository.StaticFileRepository) {
	s.(*staticFileService).staticFileRepo = repo
}

func SetSearchServiceEnqueuer(s SearchService, enqueuer SearchTaskEnqueuer) {
	s.(*searchService).searchTaskEnqueuer = enqueuer
}

func SetSearchServiceListByIDs(s SearchService, fn searchableRecordByIDsLister) {
	s.(*searchService).listSearchableByIDs = fn
}

func SetSearchServiceListRecords(s SearchService, fn searchableRecordLister) {
	s.(*searchService).listSearchableRecords = fn
}

func SetPluginRegistry(s PluginService, registry *elemoplugin.Registry) {
	s.(*pluginService).registry = registry
}

func CallPluginHost(ctx context.Context, s PluginService, pluginID string, req elemoplugin.HostRequest) (elemoplugin.HostResponse, error) {
	return s.(*pluginService).host.Call(ctx, pluginID, req)
}

func WaitPluginEvents(s PluginService) {
	s.(*pluginService).eventWG.Wait()
}

func AssertAdditiveGraph(oldS, newS *model.PluginGraphSchema) error {
	return assertAdditiveGraph(oldS, newS)
}

func ParseTypedID(raw, typ string) (model.ID, error) {
	return parseTypedID(raw, typ)
}

var (
	CtxUserID                         = ctxUserID
	RequireAction                     = requireAction
	ResolvedListScopeIDs              = resolvedListScopeIDs
	ListGrantCoversRoot               = listGrantCoversRoot
	DocumentFromRepository            = documentFromRepository
	IssueFromRepository               = issueFromRepository
	PartialIssueFromRepository        = partialIssueFromRepository
	PartialIssuesFromRepository       = partialIssuesFromRepository
	ProjectFromRepository             = projectFromRepository
	ProjectsFromRepository            = projectsFromRepository
	PartialProjectFromRepository      = partialProjectFromRepository
	NamespaceFromRepository           = namespaceFromRepository
	NamespacesFromRepository          = namespacesFromRepository
	AccessibleNamespaceFromRepository = accessibleNamespaceFromRepository
	OrganizationFromRepository        = organizationFromRepository
	RoleFromRepository                = roleFromRepository
	RolesFromRepository               = rolesFromRepository
	TodoFromRepository                = todoFromRepository
	UserFromRepository                = userFromRepository
)

const (
	DocumentFilePrefix          = documentFilePrefix
	AssignmentSyncPageSize      = assignmentSyncPageSize
	RenewEmailAddress           = renewEmailAddress
	AuthPasswordResetTemplate   = authPasswordResetTemplate
	OrganizationInviteTemplate  = organizationInviteTemplate
	SystemLicenseExpiryTemplate = systemLicenseExpiryTemplate
	UserWelcomeTemplate         = userWelcomeTemplate
)
