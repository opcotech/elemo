package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrFolderCreate       = errors.New("failed to create folder")  // the folder could not be created
	ErrFolderDelete       = errors.New("failed to delete folder")  // the folder could not be deleted
	ErrFolderRead         = errors.New("failed to read folder")    // the folder could not be retrieved
	ErrFolderUpdate       = errors.New("failed to update folder")  // the folder could not be updated
	ErrFolderNameConflict = errors.New("folder name already used") // a sibling already has this name
	ErrFolderCycle        = errors.New("folder cycle rejected")    // folder cannot be located in itself or a descendant
)

// DocumentLibrary is the organization or namespace a document or folder is scoped to.
type DocumentLibrary struct {
	ID   model.ID           `json:"id"`
	Type model.ResourceType `json:"type"`
	Name string             `json:"name"`
}

// DocumentFolder is a lean folder location on a document or nested folder.
type DocumentFolder struct {
	ID       model.ID  `json:"id"`
	Name     string    `json:"name"`
	ParentID *model.ID `json:"parent_id,omitempty"`
}

// DocumentRelation is a project or issue a document is related to.
type DocumentRelation struct {
	ID   model.ID           `json:"id"`
	Type model.ResourceType `json:"type"`
	Name string             `json:"name"`
}

// Folder represents a folder persisted by the repository.
type Folder struct {
	ID        model.ID        `json:"id"`
	Name      string          `json:"name"`
	Library   DocumentLibrary `json:"library"`
	Parent    *DocumentFolder `json:"parent"`
	CreatedBy PartialUser     `json:"created_by"`
	CreatedAt *time.Time      `json:"created_at"`
	UpdatedAt *time.Time      `json:"updated_at"`
}

// CreateFolderOpts holds the data required to create a folder.
type CreateFolderOpts struct {
	Library   model.ID
	ParentID  *model.ID
	Name      string
	CreatedBy model.ID
}

// UpdateFolderOpts holds the fields that can be updated on a folder.
type UpdateFolderOpts struct {
	Name     optional.Optional[string]
	ParentID optional.Optional[model.ID]
}

//go:generate go tool mockgen -source=folder.go -destination=mock/mock_folder_gen.go -package=mockrepo
type FolderRepository interface {
	Create(ctx context.Context, opts CreateFolderOpts) (*Folder, error)
	Get(ctx context.Context, id model.ID) (*Folder, error)
	ListForLibrary(ctx context.Context, query FolderListQuery) (Page[*Folder], error)
	Update(ctx context.Context, id model.ID, opts UpdateFolderOpts) (*Folder, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jFolderRepository is a repository for managing folders.
type Neo4jFolderRepository struct {
	*neo4jBaseRepository
}

func documentLibraryFromNode(node neo4j.Node) (DocumentLibrary, error) {
	id, err := Neo4jDecodeIDFromLabel(node)
	if err != nil {
		return DocumentLibrary{}, err
	}
	name, err := Neo4jNodeProperty[string](node, "name")
	if err != nil {
		return DocumentLibrary{}, errors.Join(ErrMalformedResult, err)
	}
	return DocumentLibrary{ID: id, Type: id.Type, Name: name}, nil
}

func documentFolderFromNode(node neo4j.Node, parentID *model.ID) (*DocumentFolder, error) {
	id, err := Neo4jDecodeID(node, model.ResourceTypeFolder)
	if err != nil {
		return nil, err
	}
	name, err := Neo4jNodeProperty[string](node, "name")
	if err != nil {
		return nil, errors.Join(ErrMalformedResult, err)
	}
	return &DocumentFolder{ID: id, Name: name, ParentID: parentID}, nil
}

func (r *Neo4jFolderRepository) scan() func(rec *neo4j.Record) (*Folder, error) {
	return func(rec *neo4j.Record) (*Folder, error) {
		node, err := Neo4jRecordNode(rec, "f")
		if err != nil {
			return nil, err
		}

		libNode, err := Neo4jRecordNode(rec, "lib")
		if err != nil {
			return nil, err
		}
		library, err := documentLibraryFromNode(libNode)
		if err != nil {
			return nil, err
		}

		createdBy, err := Neo4jRecordPartialUser(rec, "c")
		if err != nil {
			return nil, err
		}
		if createdBy == nil {
			return nil, ErrMalformedResult
		}

		folder := new(Folder)
		if err := Neo4jScanIntoStruct(&node, &folder, []string{"id", "created_by"}); err != nil {
			return nil, err
		}
		folder.ID, err = Neo4jDecodeID(node, model.ResourceTypeFolder)
		if err != nil {
			return nil, err
		}
		folder.Library = library
		folder.CreatedBy = *createdBy

		parentNode, err := Neo4jRecordOptionalNode(rec, "parent")
		if err != nil {
			return nil, err
		}
		if parentNode != nil {
			parent, err := documentFolderFromNode(*parentNode, nil)
			if err != nil {
				return nil, err
			}
			folder.Parent = parent
		}

		return folder, nil
	}
}

func (r *Neo4jFolderRepository) siblingExists(ctx context.Context, libraryID model.ID, parentID *model.ID, name, excludeID string) (bool, error) {
	params := map[string]any{
		"library_id": libraryID.String(),
		"name":       strings.ToLower(name),
		"exclude_id": excludeID,
	}

	var cypher string
	if parentID != nil {
		params["parent_id"] = parentID.String()
		cypher = `
		MATCH (:` + libraryID.Label() + ` {id: $library_id})<-[:` + EdgeKindScopedTo.String() + `]-(parent:` + model.ResourceTypeFolder.String() + ` {id: $parent_id})
		MATCH (parent)<-[:` + EdgeKindLocatedIn.String() + `]-(sibling:` + model.ResourceTypeFolder.String() + `)
		WHERE toLower(sibling.name) = $name AND sibling.id <> $exclude_id
		RETURN sibling.id AS id
		LIMIT 1`
	} else {
		cypher = `
		MATCH (:` + libraryID.Label() + ` {id: $library_id})<-[:` + EdgeKindScopedTo.String() + `]-(sibling:` + model.ResourceTypeFolder.String() + `)
		WHERE NOT (sibling)-[:` + EdgeKindLocatedIn.String() + `]->(:` + model.ResourceTypeFolder.String() + `)
		AND toLower(sibling.name) = $name AND sibling.id <> $exclude_id
		RETURN sibling.id AS id
		LIMIT 1`
	}

	rows, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, func(rec *neo4j.Record) (string, error) {
		return Neo4jParseValueFromRecord[string](rec, "id")
	})
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

func (r *Neo4jFolderRepository) wouldCycle(ctx context.Context, id, parentID model.ID) (bool, error) {
	cypher := `
	MATCH (f:` + id.Label() + ` {id: $id})
	MATCH (parent:` + parentID.Label() + ` {id: $parent_id})
	WHERE f.id = parent.id OR EXISTS { MATCH (parent)-[:` + EdgeKindLocatedIn.String() + `*]->(f) }
	RETURN f.id AS id
	LIMIT 1`
	params := map[string]any{
		"id":        id.String(),
		"parent_id": parentID.String(),
	}
	rows, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, func(rec *neo4j.Record) (string, error) {
		return Neo4jParseValueFromRecord[string](rec, "id")
	})
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

func (r *Neo4jFolderRepository) Create(ctx context.Context, opts CreateFolderOpts) (*Folder, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.FolderRepository/Create")
	defer span.End()

	exists, err := r.siblingExists(ctx, opts.Library, opts.ParentID, opts.Name, "")
	if err != nil {
		return nil, errors.Join(ErrFolderCreate, err)
	}
	if exists {
		return nil, errors.Join(ErrFolderCreate, ErrFolderNameConflict)
	}

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeFolder)

	cypher := `
	MATCH (lib:` + opts.Library.Label() + ` {id: $library_id})
	MATCH (o:` + opts.CreatedBy.Label() + ` {id: $created_by_id})`
	params := map[string]any{
		"library_id":     opts.Library.String(),
		"created_by_id":  opts.CreatedBy.String(),
		"id":             id.String(),
		"name":           opts.Name,
		"created_at":     createdAt.Format(time.RFC3339Nano),
		"scoped_rel_id":  model.NewRawID(),
		"scope_id":       model.NewRawID(),
		"created_rel_id": model.NewRawID(),
	}

	if opts.ParentID != nil {
		cypher += `
	MATCH (parent:` + opts.ParentID.Label() + ` {id: $parent_id})-[:` + EdgeKindScopedTo.String() + `]->(lib)
	CREATE
		(f:` + id.Label() + ` {
			id: $id, name: $name, created_by: $created_by_id, created_at: datetime($created_at)
		}),
		(f)-[:` + EdgeKindScopedTo.String() + ` {id: $scoped_rel_id, created_at: datetime($created_at)}]->(lib),
		(f)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(lib),
		(f)-[:` + EdgeKindLocatedIn.String() + ` {id: $located_rel_id, created_at: datetime($created_at)}]->(parent),
		(o)-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]->(f)`
		params["parent_id"] = opts.ParentID.String()
		params["located_rel_id"] = model.NewRawID()
	} else {
		cypher += `
	CREATE
		(f:` + id.Label() + ` {
			id: $id, name: $name, created_by: $created_by_id, created_at: datetime($created_at)
		}),
		(f)-[:` + EdgeKindScopedTo.String() + ` {id: $scoped_rel_id, created_at: datetime($created_at)}]->(lib),
		(f)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(lib),
		(o)-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]->(f)`
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrFolderCreate, err)
	}

	return r.Get(ctx, id)
}

func (r *Neo4jFolderRepository) Get(ctx context.Context, id model.ID) (*Folder, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.FolderRepository/Get")
	defer span.End()

	plan, err := CompileQuery(FolderGetQuery{ID: id})
	if err != nil {
		return nil, errors.Join(ErrFolderRead, err)
	}

	var folder *Folder
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		folder, _, readErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan())
		return readErr
	})
	if err != nil {
		return nil, errors.Join(ErrFolderRead, err)
	}

	return folder, nil
}

func (r *Neo4jFolderRepository) ListForLibrary(ctx context.Context, query FolderListQuery) (Page[*Folder], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.FolderRepository/ListForLibrary")
	defer span.End()

	normalized, err := query.Page.Normalize()
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderRead, err)
	}
	query.Page = normalized
	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderRead, err)
	}

	folders := make([]*Folder, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		folders, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan())
		return readErr
	})
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderRead, err)
	}

	return PaginateSlice(folders, normalized.Size, func(folder *Folder) model.ID {
		return folder.ID
	})
}

func (r *Neo4jFolderRepository) Update(ctx context.Context, id model.ID, opts UpdateFolderOpts) (*Folder, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.FolderRepository/Update")
	defer span.End()

	current, err := r.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrFolderUpdate, err)
	}

	name := current.Name
	if opts.Name.Defined && opts.Name.Value != nil {
		name = *opts.Name.Value
	}

	parentID := (*model.ID)(nil)
	if current.Parent != nil {
		parentID = &current.Parent.ID
	}
	if opts.ParentID.Defined {
		parentID = opts.ParentID.Value
	}

	if parentID != nil {
		if parentID.Type != model.ResourceTypeFolder {
			return nil, errors.Join(ErrFolderUpdate, model.ErrInvalidID)
		}
		cycles, err := r.wouldCycle(ctx, id, *parentID)
		if err != nil {
			return nil, errors.Join(ErrFolderUpdate, err)
		}
		if cycles {
			return nil, errors.Join(ErrFolderUpdate, ErrFolderCycle)
		}
	}

	exists, err := r.siblingExists(ctx, current.Library.ID, parentID, name, id.String())
	if err != nil {
		return nil, errors.Join(ErrFolderUpdate, err)
	}
	if exists {
		return nil, errors.Join(ErrFolderUpdate, ErrFolderNameConflict)
	}

	cypher := `
	MATCH (f:` + id.Label() + ` {id: $id})-[:` + EdgeKindScopedTo.String() + `]->(lib)
	OPTIONAL MATCH (f)-[loc:` + EdgeKindLocatedIn.String() + `]->()
	DELETE loc
	SET f.name = $name, f.updated_at = datetime()`
	params := map[string]any{
		"id":   id.String(),
		"name": name,
	}

	if parentID != nil {
		cypher += `
	WITH f, lib
	MATCH (parent:` + parentID.Label() + ` {id: $parent_id})-[:` + EdgeKindScopedTo.String() + `]->(lib)
	CREATE (f)-[:` + EdgeKindLocatedIn.String() + ` {id: $located_rel_id, created_at: datetime()}]->(parent)`
		params["parent_id"] = parentID.String()
		params["located_rel_id"] = model.NewRawID()
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrFolderUpdate, err)
	}

	return r.Get(ctx, id)
}

func (r *Neo4jFolderRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.FolderRepository/Delete")
	defer span.End()

	cypher := `
	MATCH (f:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (f)-[:` + EdgeKindLocatedIn.String() + `]->(parent:` + model.ResourceTypeFolder.String() + `)
	OPTIONAL MATCH (child)-[loc:` + EdgeKindLocatedIn.String() + `]->(f)
	DELETE loc
	WITH f, parent, [c IN collect(child) WHERE c IS NOT NULL] AS children
	FOREACH (c IN CASE WHEN parent IS NULL THEN [] ELSE children END |
		CREATE (c)-[:` + EdgeKindLocatedIn.String() + ` {id: randomUUID(), created_at: datetime()}]->(parent)
	)
	DETACH DELETE f`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrFolderDelete, err)
	}

	return nil
}

func clearFoldersPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeFolder.String(), pattern))
}

func clearFoldersKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearFoldersPattern(ctx, r, "Get", id.String())
}

func clearFolderAllLists(ctx context.Context, r *redisBaseRepository) error {
	return clearFoldersPattern(ctx, r, "ListForLibrary", "*")
}

func clearFolderAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	return clearDocumentsPattern(ctx, r, "*")
}

// RedisCachedFolderRepository implements caching on the FolderRepository.
type RedisCachedFolderRepository struct {
	cacheRepo  *redisBaseRepository
	folderRepo FolderRepository
}

func (r *RedisCachedFolderRepository) Create(ctx context.Context, opts CreateFolderOpts) (*Folder, error) {
	if err := clearFolderAllLists(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	return r.folderRepo.Create(ctx, opts)
}

func (r *RedisCachedFolderRepository) Get(ctx context.Context, id model.ID) (*Folder, error) {
	var folder *Folder
	var err error

	key := composeCacheKey(model.ResourceTypeFolder.String(), "Get", id.String())
	if err = r.cacheRepo.Get(ctx, key, &folder); err != nil {
		return nil, err
	}

	if folder != nil {
		return folder, nil
	}

	if folder, err = r.folderRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, folder); err != nil {
		return nil, err
	}

	return folder, nil
}

func (r *RedisCachedFolderRepository) ListForLibrary(ctx context.Context, query FolderListQuery) (Page[*Folder], error) {
	var folders Page[*Folder]
	var err error

	normalized, err := normalizedPage(query.Page)
	if err != nil {
		return Page[*Folder]{}, err
	}

	query.Page = normalized
	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*Folder]{}, err
	}
	key := plan.CacheKey(model.ResourceTypeFolder.String(), "ListForLibrary", query.LibraryID.String())
	if err = r.cacheRepo.Get(ctx, key, &folders); err != nil {
		return Page[*Folder]{}, err
	}

	if folders.Items != nil {
		return folders, nil
	}

	if folders, err = r.folderRepo.ListForLibrary(ctx, query); err != nil {
		return Page[*Folder]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, folders); err != nil {
		return Page[*Folder]{}, err
	}

	return folders, nil
}

func (r *RedisCachedFolderRepository) Update(ctx context.Context, id model.ID, opts UpdateFolderOpts) (*Folder, error) {
	folder, err := r.folderRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeFolder.String(), "Get", id.String())
	if err = r.cacheRepo.Set(ctx, key, folder); err != nil {
		return nil, err
	}

	if err := clearFolderAllLists(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return folder, nil
}

func (r *RedisCachedFolderRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearFoldersKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}
	if err := clearFolderAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearFolderAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}
	return r.folderRepo.Delete(ctx, id)
}

// NewNeo4jFolderRepository creates a new folder repository.
func NewNeo4jFolderRepository(opts ...Neo4jRepositoryOption) (*Neo4jFolderRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jFolderRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

// NewCachedFolderRepository returns a new cached FolderRepository.
func NewCachedFolderRepository(repo FolderRepository, opts ...RedisRepositoryOption) (*RedisCachedFolderRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedFolderRepository{
		cacheRepo:  r,
		folderRepo: repo,
	}, nil
}
