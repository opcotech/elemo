package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/repository"
)

const (
	DefaultSearchPageSize = 20
	MinSearchPageSize     = 1
	MaxSearchPageSize     = 100
	maxSearchFetchRounds  = 8
)

var searchableResourceTypes = []model.ResourceType{
	model.ResourceTypeOrganization,
	model.ResourceTypeNamespace,
	model.ResourceTypeProject,
	model.ResourceTypeIssue,
	model.ResourceTypeDocument,
}

// SearchReindexSources holds the graph stores Reindex walks. They are
// call-scoped so SearchService construction does not need them. Nil
// domain repositories are skipped with a warning.
type SearchReindexSources struct {
	DB           *repository.Neo4jDatabase
	Organization repository.OrganizationRepository
	Namespace    repository.NamespaceRepository
	Project      repository.ProjectRepository
	Issue        repository.IssueRepository
	Document     repository.DocumentRepository
}

// IndexInput is the searchable projection of a persisted resource. Ancestry is
// filled by SearchService from IN_SCOPE_OF.
type IndexInput struct {
	ID        model.ID
	Title     string
	Content   string
	Key       string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// SearchQuery is a permission-aware search request.
type SearchQuery struct {
	Text           string
	Types          []model.ResourceType
	OrganizationID *model.ID
	NamespaceID    *model.ID
	ProjectID      *model.ID
	PageSize       int
	PageToken      *string
}

// SearchResult is a public hit. It never includes scope ancestry or
// Meilisearch ranking fields.
type SearchResult struct {
	ID             model.ID
	Type           model.ResourceType
	Title          string
	Subtitle       string
	Key            string
	OrganizationID *model.ID
	NamespaceID    *model.ID
	ProjectID      *model.ID
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

type searchPageToken struct {
	Offset int64  `json:"o"`
	Hash   string `json:"h"`
}

//go:generate go tool mockgen -destination=search_mock_gen.go -package=service -mock_names SearchService=MockSearchService . SearchService
type SearchService interface {
	// Index upserts a searchable projection. Ancestry is resolved from Neo4j.
	Index(ctx context.Context, input IndexInput) error
	// Delete removes one document by resource ID.
	Delete(ctx context.Context, id model.ID) error
	// DeleteByScope removes every document whose ancestry includes scope.
	DeleteByScope(ctx context.Context, scope model.ID) error
	// Search returns a page of hits the context user may read.
	Search(ctx context.Context, q SearchQuery) (Page[*SearchResult], error)
	// Reindex walks Neo4j and rebuilds the search index.
	Reindex(ctx context.Context, sources SearchReindexSources) error
}

type searchService struct {
	*baseService
	searchRepo repository.SearchRepository
}

func unixSeconds(t *time.Time) int64 {
	if t == nil || t.IsZero() {
		return 0
	}
	return t.Unix()
}

func parseOptionalComposite(value string) *model.ID {
	if value == "" {
		return nil
	}
	id, err := model.ParseCompositeID(value)
	if err != nil {
		return nil
	}
	return &id
}

func searchResultSubtitle(doc repository.SearchDocument) string {
	switch doc.Type {
	case model.ResourceTypeIssue.String(), model.ResourceTypeProject.String():
		return doc.Key
	case model.ResourceTypeDocument.String():
		return doc.Content
	default:
		return doc.Content
	}
}

func searchDocumentToResult(doc repository.SearchDocument) (*SearchResult, error) {
	id, err := model.ParseSearchKey(doc.ID)
	if err != nil {
		return nil, err
	}
	createdAt := time.Unix(doc.CreatedAt, 0).UTC()
	var updatedAt *time.Time
	if doc.UpdatedAt > 0 {
		t := time.Unix(doc.UpdatedAt, 0).UTC()
		updatedAt = &t
	}
	return &SearchResult{
		ID:             id,
		Type:           id.Type,
		Title:          doc.Title,
		Subtitle:       searchResultSubtitle(doc),
		Key:            doc.Key,
		OrganizationID: parseOptionalComposite(doc.OrganizationID),
		NamespaceID:    parseOptionalComposite(doc.NamespaceID),
		ProjectID:      parseOptionalComposite(doc.ProjectID),
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

func (s *searchService) Index(ctx context.Context, input IndexInput) error {
	ctx, span := s.tracer.Start(ctx, "service.searchService/Index")
	defer span.End()

	if err := input.ID.Validate(); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}

	ancestry, err := s.permissionService.ListScopeAncestry(ctx, input.ID)
	if err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	if len(ancestry) == 0 {
		ancestry = []model.ID{input.ID}
	}

	scopeIDs := make([]string, 0, len(ancestry))
	doc := repository.SearchDocument{
		ID:        input.ID.SearchKey(),
		Type:      input.ID.Type.String(),
		Title:     input.Title,
		Content:   input.Content,
		Key:       input.Key,
		ScopeIDs:  make([]string, 0, len(ancestry)),
		CreatedAt: unixSeconds(input.CreatedAt),
		UpdatedAt: unixSeconds(input.UpdatedAt),
	}
	for _, scope := range ancestry {
		scopeIDs = append(scopeIDs, scope.Composite())
		switch scope.Type {
		case model.ResourceTypeOrganization:
			doc.OrganizationID = scope.Composite()
		case model.ResourceTypeNamespace:
			doc.NamespaceID = scope.Composite()
		case model.ResourceTypeProject:
			doc.ProjectID = scope.Composite()
		}
	}
	doc.ScopeIDs = scopeIDs

	if err := s.searchRepo.Upsert(ctx, doc); err != nil {
		return errors.Join(ErrSearchIndex, err)
	}
	return nil
}

func (s *searchService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.searchService/Delete")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	if err := s.searchRepo.Delete(ctx, id.SearchKey()); err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	return nil
}

func (s *searchService) DeleteByScope(ctx context.Context, scope model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.searchService/DeleteByScope")
	defer span.End()

	if err := scope.Validate(); err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	if err := s.searchRepo.DeleteByScope(ctx, scope.Composite()); err != nil {
		return errors.Join(ErrSearchDelete, err)
	}
	return nil
}

func (s *searchService) Search(ctx context.Context, q SearchQuery) (Page[*SearchResult], error) {
	ctx, span := s.tracer.Start(ctx, "service.searchService/Search")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return Page[*SearchResult]{}, errors.Join(ErrSearchGet, ErrNoUser)
	}

	types, err := normalizeSearchTypes(q.Types)
	if err != nil {
		return Page[*SearchResult]{}, errors.Join(ErrSearchGet, err)
	}

	pageSize := q.PageSize
	if pageSize == 0 {
		pageSize = DefaultSearchPageSize
	}
	if pageSize < MinSearchPageSize || pageSize > MaxSearchPageSize {
		return Page[*SearchResult]{}, errors.Join(ErrSearchGet, repository.ErrInvalidPageSize)
	}

	fingerprint := searchQueryFingerprint(q, types)
	offset := int64(0)
	if q.PageToken != nil && *q.PageToken != "" {
		token, err := decodeSearchPageToken(*q.PageToken)
		if err != nil {
			return Page[*SearchResult]{}, errors.Join(ErrSearchGet, err)
		}
		if token.Hash != fingerprint {
			return Page[*SearchResult]{}, errors.Join(ErrSearchGet, repository.ErrInvalidCursor)
		}
		offset = token.Offset
	}

	typeFilters := make([]repository.SearchTypeFilter, 0, len(types))
	for _, rt := range types {
		action, ok := model.ReadActionFor(rt)
		if !ok {
			continue
		}
		scopes, err := s.permissionService.ListGrantScopes(ctx, userID, action)
		if err != nil {
			return Page[*SearchResult]{}, errors.Join(ErrSearchGet, err)
		}
		if len(scopes) == 0 {
			continue
		}
		scopeIDs := make([]string, len(scopes))
		for i, scope := range scopes {
			scopeIDs[i] = scope.Composite()
		}
		typeFilters = append(typeFilters, repository.SearchTypeFilter{
			Type:     rt.String(),
			ScopeIDs: scopeIDs,
		})
	}
	if len(typeFilters) == 0 {
		return repository.EmptyPage[*SearchResult](), nil
	}

	repoQuery := repository.SearchQuery{
		Text:        q.Text,
		TypeFilters: typeFilters,
		Offset:      offset,
		Limit:       int64(pageSize),
	}
	if q.OrganizationID != nil {
		if err := q.OrganizationID.Validate(); err != nil {
			return Page[*SearchResult]{}, errors.Join(ErrSearchGet, err)
		}
		repoQuery.OrganizationID = q.OrganizationID.Composite()
	}
	if q.NamespaceID != nil {
		if err := q.NamespaceID.Validate(); err != nil {
			return Page[*SearchResult]{}, errors.Join(ErrSearchGet, err)
		}
		repoQuery.NamespaceID = q.NamespaceID.Composite()
	}
	if q.ProjectID != nil {
		if err := q.ProjectID.Validate(); err != nil {
			return Page[*SearchResult]{}, errors.Join(ErrSearchGet, err)
		}
		repoQuery.ProjectID = q.ProjectID.Composite()
	}

	collected := make([]*SearchResult, 0, pageSize)
	nextOffset := offset
	exhausted := false

	for round := 0; round < maxSearchFetchRounds && len(collected) < pageSize; round++ {
		need := pageSize - len(collected)
		fetch := int64(need * 2)
		if fetch < int64(need) {
			fetch = int64(need)
		}
		repoQuery.Offset = nextOffset
		repoQuery.Limit = fetch

		hits, err := s.searchRepo.Search(ctx, repoQuery)
		if err != nil {
			return Page[*SearchResult]{}, errors.Join(ErrSearchGet, err)
		}
		if len(hits.Documents) == 0 {
			exhausted = true
			break
		}

		for _, doc := range hits.Documents {
			nextOffset++
			result, err := searchDocumentToResult(doc)
			if err != nil {
				continue
			}
			action, ok := model.ReadActionFor(result.Type)
			if !ok {
				continue
			}
			allowed, err := s.permissionService.Has(ctx, userID, result.ID, action)
			if err != nil || !allowed {
				continue
			}
			collected = append(collected, result)
			if len(collected) == pageSize {
				break
			}
		}

		if int64(len(hits.Documents)) < fetch {
			exhausted = true
			break
		}
	}

	hasMore := !exhausted && len(collected) == pageSize
	page := Page[*SearchResult]{
		Items: collected,
		PageInfo: PageInfo{
			HasMore: hasMore,
		},
	}
	if hasMore {
		token, err := encodeSearchPageToken(searchPageToken{Offset: nextOffset, Hash: fingerprint})
		if err != nil {
			return Page[*SearchResult]{}, errors.Join(ErrSearchGet, err)
		}
		page.PageInfo.NextPageToken = &token
	}
	if page.Items == nil {
		page.Items = []*SearchResult{}
	}
	return page, nil
}

func (s *searchService) Reindex(ctx context.Context, sources SearchReindexSources) error {
	ctx, span := s.tracer.Start(ctx, "service.searchService/Reindex")
	defer span.End()

	if sources.DB == nil {
		return errors.Join(ErrSearchReindex, repository.ErrNoDriver)
	}

	steps := []struct {
		skip bool
		kind model.ResourceType
		run  func() error
	}{
		{sources.Organization == nil, model.ResourceTypeOrganization, func() error {
			return s.reindexOrganizations(ctx, sources)
		}},
		{sources.Namespace == nil, model.ResourceTypeNamespace, func() error {
			return s.reindexNamespaces(ctx, sources)
		}},
		{sources.Project == nil, model.ResourceTypeProject, func() error {
			return s.reindexProjects(ctx, sources)
		}},
		{sources.Issue == nil, model.ResourceTypeIssue, func() error {
			return s.reindexIssues(ctx, sources)
		}},
		{sources.Document == nil, model.ResourceTypeDocument, func() error {
			return s.reindexDocuments(ctx, sources)
		}},
	}
	for _, step := range steps {
		if step.skip {
			s.logger.Warn(ctx, "skipping nil search reindex source", log.WithKind(step.kind.String()))
			continue
		}
		if err := step.run(); err != nil {
			return err
		}
	}
	return nil
}

func (s *searchService) reindexOrganizations(ctx context.Context, sources SearchReindexSources) error {
	ids, err := repository.ListSearchableIDs(ctx, sources.DB, model.ResourceTypeOrganization)
	if err != nil {
		return errors.Join(ErrSearchReindex, err)
	}
	for _, id := range ids {
		org, err := sources.Organization.Get(ctx, id, repository.OrganizationDetailProjection())
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				continue
			}
			return errors.Join(ErrSearchReindex, err)
		}
		if err := s.Index(ctx, IndexInput{
			ID:        org.ID,
			Title:     org.Name,
			CreatedAt: org.CreatedAt,
			UpdatedAt: org.UpdatedAt,
		}); err != nil {
			return errors.Join(ErrSearchReindex, err)
		}
	}
	return nil
}

func (s *searchService) reindexNamespaces(ctx context.Context, sources SearchReindexSources) error {
	ids, err := repository.ListSearchableIDs(ctx, sources.DB, model.ResourceTypeNamespace)
	if err != nil {
		return errors.Join(ErrSearchReindex, err)
	}
	for _, id := range ids {
		ns, err := sources.Namespace.Get(ctx, id, repository.NamespaceDetailProjection())
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				continue
			}
			return errors.Join(ErrSearchReindex, err)
		}
		if err := s.Index(ctx, IndexInput{
			ID:        ns.ID,
			Title:     ns.Name,
			Content:   ns.Description,
			CreatedAt: ns.CreatedAt,
			UpdatedAt: ns.UpdatedAt,
		}); err != nil {
			return errors.Join(ErrSearchReindex, err)
		}
	}
	return nil
}

func (s *searchService) reindexProjects(ctx context.Context, sources SearchReindexSources) error {
	ids, err := repository.ListSearchableIDs(ctx, sources.DB, model.ResourceTypeProject)
	if err != nil {
		return errors.Join(ErrSearchReindex, err)
	}
	for _, id := range ids {
		project, err := sources.Project.Get(ctx, id, repository.ProjectDetailProjection())
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				continue
			}
			return errors.Join(ErrSearchReindex, err)
		}
		if err := s.Index(ctx, IndexInput{
			ID:        project.ID,
			Title:     project.Name,
			Content:   project.Description,
			Key:       project.Key,
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdatedAt,
		}); err != nil {
			return errors.Join(ErrSearchReindex, err)
		}
	}
	return nil
}

func (s *searchService) reindexIssues(ctx context.Context, sources SearchReindexSources) error {
	ids, err := repository.ListSearchableIDs(ctx, sources.DB, model.ResourceTypeIssue)
	if err != nil {
		return errors.Join(ErrSearchReindex, err)
	}
	for _, id := range ids {
		issue, err := sources.Issue.Get(ctx, id, repository.IssueDetailProjection())
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				continue
			}
			return errors.Join(ErrSearchReindex, err)
		}
		if err := s.Index(ctx, IndexInput{
			ID:        issue.ID,
			Title:     issue.Title,
			Content:   issue.Description,
			Key:       issue.Key,
			CreatedAt: issue.CreatedAt,
			UpdatedAt: issue.UpdatedAt,
		}); err != nil {
			return errors.Join(ErrSearchReindex, err)
		}
	}
	return nil
}

func (s *searchService) reindexDocuments(ctx context.Context, sources SearchReindexSources) error {
	ids, err := repository.ListSearchableIDs(ctx, sources.DB, model.ResourceTypeDocument)
	if err != nil {
		return errors.Join(ErrSearchReindex, err)
	}
	for _, id := range ids {
		doc, err := sources.Document.Get(ctx, id, repository.DocumentDetailProjection())
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				continue
			}
			return errors.Join(ErrSearchReindex, err)
		}
		if err := s.Index(ctx, IndexInput{
			ID:        doc.ID,
			Title:     doc.Title,
			Content:   doc.Excerpt,
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}); err != nil {
			return errors.Join(ErrSearchReindex, err)
		}
	}
	return nil
}

func normalizeSearchTypes(types []model.ResourceType) ([]model.ResourceType, error) {
	if len(types) == 0 {
		out := make([]model.ResourceType, len(searchableResourceTypes))
		copy(out, searchableResourceTypes)
		return out, nil
	}
	seen := make(map[model.ResourceType]struct{}, len(types))
	out := make([]model.ResourceType, 0, len(types))
	allowed := make(map[model.ResourceType]struct{}, len(searchableResourceTypes))
	for _, rt := range searchableResourceTypes {
		allowed[rt] = struct{}{}
	}
	for _, rt := range types {
		if _, ok := allowed[rt]; !ok {
			return nil, model.ErrInvalidResourceType
		}
		if _, ok := seen[rt]; ok {
			continue
		}
		seen[rt] = struct{}{}
		out = append(out, rt)
	}
	return out, nil
}

func searchQueryFingerprint(q SearchQuery, types []model.ResourceType) string {
	labels := make([]string, len(types))
	for i, rt := range types {
		labels[i] = rt.String()
	}
	sort.Strings(labels)
	payload := struct {
		Text  string
		Types []string
		Org   string
		NS    string
		Proj  string
	}{
		Text:  q.Text,
		Types: labels,
	}
	if q.OrganizationID != nil {
		payload.Org = q.OrganizationID.Composite()
	}
	if q.NamespaceID != nil {
		payload.NS = q.NamespaceID.Composite()
	}
	if q.ProjectID != nil {
		payload.Proj = q.ProjectID.Composite()
	}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func encodeSearchPageToken(token searchPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", errors.Join(repository.ErrInvalidCursor, err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeSearchPageToken(value string) (searchPageToken, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return searchPageToken{}, errors.Join(repository.ErrInvalidCursor, err)
	}
	var token searchPageToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return searchPageToken{}, errors.Join(repository.ErrInvalidCursor, err)
	}
	if token.Offset < 0 || token.Hash == "" {
		return searchPageToken{}, repository.ErrInvalidCursor
	}
	return token, nil
}

func NewSearchService(searchRepo repository.SearchRepository, opts ...Option) (SearchService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &searchService{
		baseService: s,
		searchRepo:  searchRepo,
	}
	if svc.searchRepo == nil {
		return nil, ErrNoSearchRepository
	}
	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}
	return svc, nil
}
