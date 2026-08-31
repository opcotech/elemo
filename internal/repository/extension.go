package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
)

var (
	ErrExtensionCreate = errors.New("failed to create extension")
	ErrExtensionRead   = errors.New("failed to read extension")
	ErrExtensionUpdate = errors.New("failed to update extension")
	ErrExtensionDelete = errors.New("failed to delete extension")
	ErrExtensionParent = errors.New("extension parent not found")
)

type CreateExtensionOpts struct {
	PluginID   string
	Kind       string
	Parent     model.ID
	Properties map[string]any
}

type UpdateExtensionOpts struct {
	Properties map[string]any
}

type ListExtensionFilter struct {
	PluginID string
	Kind     string
	Scope    model.ID
	Equals   map[string]any
	Page     CursorPage
}

type CreateExtensionRelationOpts struct {
	PluginID string
	Kind     string
	From     model.ID
	To       model.ID
}

type MoveExtensionOpts struct {
	PluginID      string
	ID            model.ID
	Parent        model.ID
	RelationTypes []string
}

//go:generate go tool mockgen -source=extension.go -destination=mock/mock_extension_gen.go -package=mockrepo
type ExtensionRepository interface {
	Create(ctx context.Context, opts CreateExtensionOpts) (*model.Extension, error)
	Get(ctx context.Context, pluginID string, id model.ID) (*model.Extension, error)
	Update(ctx context.Context, pluginID string, id model.ID, opts UpdateExtensionOpts) (*model.Extension, error)
	Move(ctx context.Context, opts MoveExtensionOpts) (*model.Extension, error)
	Delete(ctx context.Context, pluginID string, id model.ID) error
	List(ctx context.Context, filter ListExtensionFilter) (Page[*model.Extension], error)

	CreateRelation(ctx context.Context, opts CreateExtensionRelationOpts) (*model.ExtensionRelation, error)
	DeleteRelation(ctx context.Context, pluginID, relID string) error
	ListRelations(ctx context.Context, pluginID, kind string, node model.ID, direction model.PluginGraphRelationDirection, page CursorPage) (Page[*model.ExtensionRelation], error)
	CountRelations(ctx context.Context, pluginID, kind string, from, to model.ID) (int64, int64, error)

	DeleteByPlugin(ctx context.Context, pluginID string) error
}

type Neo4jExtensionRepository struct {
	*neo4jBaseRepository
}

func NewExtensionRepository(opts ...Neo4jRepositoryOption) (*Neo4jExtensionRepository, error) {
	base, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}
	return &Neo4jExtensionRepository{neo4jBaseRepository: base}, nil
}

func (r *Neo4jExtensionRepository) scanNode() func(rec *neo4j.Record) (*model.Extension, error) {
	return func(rec *neo4j.Record) (*model.Extension, error) {
		node, _, err := neo4j.GetRecordValue[neo4j.Node](rec, "e")
		if err != nil {
			return nil, err
		}
		ext, err := extensionFromNode(node)
		if err != nil {
			return nil, err
		}
		parent, _, parentErr := neo4j.GetRecordValue[neo4j.Node](rec, "parent")
		if parentErr == nil {
			if id, idErr := idFromNeo4jNode(parent); idErr == nil {
				ext.Parent = &id
			}
		}
		return ext, nil
	}
}

func idFromNeo4jNode(node neo4j.Node) (model.ID, error) {
	props := node.GetProperties()
	idStr, _ := props["id"].(string)
	if idStr == "" {
		return model.ID{}, errors.New("missing node id")
	}
	for _, label := range node.Labels {
		id, err := model.NewIDFromString(idStr, label)
		if err == nil {
			return id, nil
		}
	}
	return model.ID{}, fmt.Errorf("unknown node type for %s", idStr)
}

func extensionFromNode(node neo4j.Node) (*model.Extension, error) {
	props := node.GetProperties()
	idStr, _ := props["id"].(string)
	pluginID, _ := props["plugin_id"].(string)
	kind, _ := props["kind"].(string)
	id, err := model.NewIDFromString(idStr, model.ResourceTypeExtension.String())
	if err != nil {
		return nil, err
	}
	fields := make(map[string]any, len(props))
	for k, v := range props {
		if model.IsReservedGraphProperty(k) {
			continue
		}
		fields[k] = v
	}
	ext := &model.Extension{
		ID:         id,
		PluginID:   pluginID,
		Kind:       kind,
		Properties: fields,
	}
	if t, ok := props["created_at"].(time.Time); ok {
		ext.CreatedAt = &t
	}
	if t, ok := props["updated_at"].(time.Time); ok {
		ext.UpdatedAt = &t
	}
	return ext, nil
}

func (r *Neo4jExtensionRepository) Create(ctx context.Context, opts CreateExtensionOpts) (*model.Extension, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/Create")
	defer span.End()

	ext, err := model.NewExtension(opts.PluginID, opts.Kind, opts.Properties)
	if err != nil {
		return nil, errors.Join(ErrExtensionCreate, err)
	}
	now := time.Now().UTC()
	ext.CreatedAt = &now

	props := map[string]any{
		"id":         ext.ID.String(),
		"plugin_id":  ext.PluginID,
		"kind":       ext.Kind,
		"created_at": now.Format(time.RFC3339Nano),
	}
	for k, v := range opts.Properties {
		if model.IsReservedGraphProperty(k) {
			return nil, errors.Join(ErrExtensionCreate, fmt.Errorf("property %s is reserved", k))
		}
		props[k] = v
	}

	cypher := `
	MATCH (parent:` + opts.Parent.Label() + ` {id: $parent_id})
	CREATE (e:` + model.ResourceTypeExtension.String() + `)
	SET e = $props
	SET e.created_at = datetime($created_at)
	CREATE (e)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(parent)
	RETURN e, parent`

	params := map[string]any{
		"parent_id":  opts.Parent.String(),
		"props":      props,
		"created_at": now.Format(time.RFC3339Nano),
		"scope_id":   model.NewRawID(),
	}

	created, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scanNode())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errors.Join(ErrExtensionCreate, ErrExtensionParent)
		}
		return nil, errors.Join(ErrExtensionCreate, err)
	}
	return created, nil
}

func (r *Neo4jExtensionRepository) Get(ctx context.Context, pluginID string, id model.ID) (*model.Extension, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/Get")
	defer span.End()

	cypher := `
	MATCH (e:` + model.ResourceTypeExtension.String() + ` {id: $id, plugin_id: $plugin_id})
	OPTIONAL MATCH (e)-[:` + EdgeKindInScopeOf.String() + `]->(parent)
	RETURN e, parent`
	params := map[string]any{"id": id.String(), "plugin_id": pluginID}
	ext, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scanNode())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrExtensionRead, err)
	}
	return ext, nil
}

func (r *Neo4jExtensionRepository) Update(
	ctx context.Context,
	pluginID string,
	id model.ID,
	opts UpdateExtensionOpts,
) (*model.Extension, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/Update")
	defer span.End()

	props := make(map[string]any, len(opts.Properties))
	for k, v := range opts.Properties {
		if model.IsReservedGraphProperty(k) {
			return nil, errors.Join(ErrExtensionUpdate, fmt.Errorf("property %s is reserved", k))
		}
		props[k] = v
	}
	now := time.Now().UTC()
	cypher := `
	MATCH (e:` + model.ResourceTypeExtension.String() + ` {id: $id, plugin_id: $plugin_id})
	SET e += $props
	SET e.updated_at = datetime($updated_at)
	WITH e
	OPTIONAL MATCH (e)-[:` + EdgeKindInScopeOf.String() + `]->(parent)
	RETURN e, parent`
	params := map[string]any{
		"id":         id.String(),
		"plugin_id":  pluginID,
		"props":      props,
		"updated_at": now.Format(time.RFC3339Nano),
	}
	ext, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scanNode())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(ErrExtensionUpdate, err)
	}
	return ext, nil
}

func (r *Neo4jExtensionRepository) Delete(ctx context.Context, pluginID string, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/Delete")
	defer span.End()

	cypher := `
	MATCH (e:` + model.ResourceTypeExtension.String() + ` {id: $id, plugin_id: $plugin_id})
	DETACH DELETE e`
	err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, map[string]any{
		"id":        id.String(),
		"plugin_id": pluginID,
	})
	if err != nil {
		return errors.Join(ErrExtensionDelete, err)
	}
	return nil
}

func (r *Neo4jExtensionRepository) Move(ctx context.Context, opts MoveExtensionOpts) (*model.Extension, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/Move")
	defer span.End()

	now := time.Now().UTC()
	params := map[string]any{
		"id":         opts.ID.String(),
		"plugin_id":  opts.PluginID,
		"parent_id":  opts.Parent.String(),
		"created_at": now.Format(time.RFC3339Nano),
		"scope_id":   model.NewRawID(),
		"from_type":  model.ResourceTypeExtension.String(),
		"to_type":    opts.Parent.Label(),
	}

	var b strings.Builder
	b.WriteString(`
	MATCH (e:` + model.ResourceTypeExtension.String() + ` {id: $id, plugin_id: $plugin_id})
	MATCH (e)-[old:` + EdgeKindInScopeOf.String() + `]->()
	MATCH (parent:` + opts.Parent.Label() + ` {id: $parent_id})
	DELETE old
	CREATE (e)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(parent)
	`)
	for i, relType := range opts.RelationTypes {
		if !strings.HasPrefix(relType, "EXT__") {
			return nil, errors.Join(ErrExtensionUpdate, fmt.Errorf("invalid relation type"))
		}
		relIDKey := fmt.Sprintf("rel_id_%d", i)
		params[relIDKey] = model.NewRawID()
		fmt.Fprintf(&b, `
	WITH e, parent
	OPTIONAL MATCH (e)-[r%d:%s]->()
	DELETE r%d
	WITH e, parent
	CREATE (e)-[:%s {id: $%s, created_at: datetime($created_at), from_type: $from_type, to_type: $to_type}]->(parent)
`, i, relType, i, relType, relIDKey)
	}
	b.WriteString("\n	RETURN e, parent")

	ext, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, b.String(), params, r.scanNode())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errors.Join(ErrExtensionUpdate, ErrExtensionParent)
		}
		return nil, errors.Join(ErrExtensionUpdate, err)
	}
	return ext, nil
}

func (r *Neo4jExtensionRepository) List(ctx context.Context, filter ListExtensionFilter) (Page[*model.Extension], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/List")
	defer span.End()

	normalized, err := filter.Page.Normalize()
	if err != nil {
		return Page[*model.Extension]{}, errors.Join(ErrExtensionRead, err)
	}

	params := map[string]any{
		"plugin_id": filter.PluginID,
		"kind":      filter.Kind,
		"scope_id":  filter.Scope.String(),
		"limit":     normalized.FetchLimit(),
	}
	var predicates []string
	if err := ApplyCursorParams(params, normalized.Token); err != nil {
		return Page[*model.Extension]{}, errors.Join(ErrExtensionRead, err)
	}
	if normalized.Token != nil {
		predicates = append(predicates, "e.id > $cursor_id")
	}
	if len(filter.Equals) > 0 {
		for k, v := range filter.Equals {
			if model.IsReservedGraphProperty(k) {
				continue
			}
			predicates = append(predicates, "e[$prop] = $value")
			params["prop"] = k
			params["value"] = v
			break
		}
	}
	where := ""
	if len(predicates) > 0 {
		where = "WHERE " + strings.Join(predicates, " AND ")
	}

	cypher := `
	MATCH (e:` + model.ResourceTypeExtension.String() + ` {plugin_id: $plugin_id, kind: $kind})
	MATCH (e)-[:` + EdgeKindInScopeOf.String() + `*0..]->(scope {id: $scope_id})
	OPTIONAL MATCH (e)-[:` + EdgeKindInScopeOf.String() + `]->(parent)
	` + where + `
	RETURN e, parent
	ORDER BY e.id
	LIMIT $limit`

	items, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scanNode())
	if err != nil {
		return Page[*model.Extension]{}, errors.Join(ErrExtensionRead, err)
	}
	page, err := PaginateSlice(items, normalized.Size, func(ext *model.Extension) model.ID {
		return ext.ID
	})
	if err != nil {
		return Page[*model.Extension]{}, errors.Join(ErrExtensionRead, err)
	}
	return page, nil
}

func (r *Neo4jExtensionRepository) CreateRelation(
	ctx context.Context,
	opts CreateExtensionRelationOpts,
) (*model.ExtensionRelation, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/CreateRelation")
	defer span.End()

	relType, err := elemoplugin.RelationType(opts.PluginID, opts.Kind)
	if err != nil {
		return nil, errors.Join(ErrExtensionCreate, err)
	}
	now := time.Now().UTC()
	relID := model.NewRawID()
	cypher := `
	MATCH (a:` + opts.From.Label() + ` {id: $from_id})
	MATCH (b:` + opts.To.Label() + ` {id: $to_id})
	CREATE (a)-[r:` + relType + ` {
		id: $id,
		created_at: datetime($created_at),
		from_type: $from_type,
		to_type: $to_type
	}]->(b)
	RETURN r.id AS id`

	params := map[string]any{
		"from_id":    opts.From.String(),
		"to_id":      opts.To.String(),
		"from_type":  opts.From.Label(),
		"to_type":    opts.To.Label(),
		"id":         relID,
		"created_at": now.Format(time.RFC3339Nano),
	}
	created, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*model.ExtensionRelation, error) {
		id, _ := rec.Values[0].(string)
		return &model.ExtensionRelation{
			ID:        id,
			Kind:      opts.Kind,
			From:      opts.From,
			To:        opts.To,
			CreatedAt: &now,
		}, nil
	})
	if err != nil {
		return nil, errors.Join(ErrExtensionCreate, err)
	}
	return created, nil
}

func (r *Neo4jExtensionRepository) DeleteRelation(ctx context.Context, pluginID, relID string) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/DeleteRelation")
	defer span.End()

	prefix := elemoplugin.RelationPrefix(pluginID)
	cypher := `
	MATCH ()-[r]->()
	WHERE r.id = $id AND type(r) STARTS WITH $prefix
	DELETE r`
	err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, map[string]any{
		"id":     relID,
		"prefix": prefix,
	})
	if err != nil {
		return errors.Join(ErrExtensionDelete, err)
	}
	return nil
}

func (r *Neo4jExtensionRepository) ListRelations(
	ctx context.Context,
	pluginID, kind string,
	node model.ID,
	direction model.PluginGraphRelationDirection,
	page CursorPage,
) (Page[*model.ExtensionRelation], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/ListRelations")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*model.ExtensionRelation]{}, errors.Join(ErrExtensionRead, err)
	}
	relType, err := elemoplugin.RelationType(pluginID, kind)
	if err != nil {
		return Page[*model.ExtensionRelation]{}, errors.Join(ErrExtensionRead, err)
	}
	if direction == model.PluginGraphRelationDirectionUnknown {
		direction = model.PluginGraphRelationDirectionOutgoing
	}
	if !direction.IsAPluginGraphRelationDirection() || direction == model.PluginGraphRelationDirectionUnknown {
		return Page[*model.ExtensionRelation]{}, errors.Join(ErrExtensionRead, fmt.Errorf("invalid relation direction"))
	}

	match := `(n:` + node.Label() + ` {id: $id})-[r:` + relType + `]->(m)`
	switch direction {
	case model.PluginGraphRelationDirectionIncoming:
		match = `(n:` + node.Label() + ` {id: $id})<-[r:` + relType + `]-(m)`
	case model.PluginGraphRelationDirectionBoth:
		match = `(n:` + node.Label() + ` {id: $id})-[r:` + relType + `]-(m)`
	}

	cypher := `
	MATCH ` + match + `
	RETURN r.id AS id, startNode(r).id AS from_id, r.from_type AS from_type,
		endNode(r).id AS to_id, r.to_type AS to_type, r.created_at AS created_at
	ORDER BY r.id
	LIMIT $limit`
	params := map[string]any{"id": node.String(), "limit": normalized.Size + 1}
	items, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, scanExtensionRelation(kind))
	if err != nil {
		return Page[*model.ExtensionRelation]{}, errors.Join(ErrExtensionRead, err)
	}
	out, err := PaginateSlice(items, normalized.Size, func(rel *model.ExtensionRelation) model.ID {
		id, err := model.NewIDFromString(rel.ID, model.ResourceTypeExtension.String())
		if err != nil {
			return rel.From
		}
		return id
	})
	if err != nil {
		return Page[*model.ExtensionRelation]{}, errors.Join(ErrExtensionRead, err)
	}
	return out, nil
}

func scanExtensionRelation(kind string) func(rec *neo4j.Record) (*model.ExtensionRelation, error) {
	return func(rec *neo4j.Record) (*model.ExtensionRelation, error) {
		id, _ := rec.Get("id")
		fromID, _ := rec.Get("from_id")
		fromType, _ := rec.Get("from_type")
		toID, _ := rec.Get("to_id")
		toType, _ := rec.Get("to_type")
		from, err := model.NewIDFromString(fmt.Sprint(fromID), fmt.Sprint(fromType))
		if err != nil {
			return nil, err
		}
		to, err := model.NewIDFromString(fmt.Sprint(toID), fmt.Sprint(toType))
		if err != nil {
			return nil, err
		}
		rel := &model.ExtensionRelation{ID: fmt.Sprint(id), Kind: kind, From: from, To: to}
		if t, ok := rec.Values[5].(time.Time); ok {
			rel.CreatedAt = &t
		}
		return rel, nil
	}
}

func (r *Neo4jExtensionRepository) CountRelations(
	ctx context.Context,
	pluginID, kind string,
	from, to model.ID,
) (int64, int64, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/CountRelations")
	defer span.End()

	relType, err := elemoplugin.RelationType(pluginID, kind)
	if err != nil {
		return 0, 0, errors.Join(ErrExtensionRead, err)
	}
	cypher := `
	OPTIONAL MATCH (a:` + from.Label() + ` {id: $from_id})-[out:` + relType + `]->()
	WITH count(out) AS outgoing
	OPTIONAL MATCH ()-[in:` + relType + `]->(b:` + to.Label() + ` {id: $to_id})
	RETURN outgoing, count(in) AS incoming`
	params := map[string]any{"from_id": from.String(), "to_id": to.String()}
	counts, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, scanRelationCounts)
	if err != nil {
		return 0, 0, errors.Join(ErrExtensionRead, err)
	}
	return counts.Outgoing, counts.Incoming, nil
}

type relationCounts struct {
	Outgoing int64
	Incoming int64
}

func scanRelationCounts(rec *neo4j.Record) (*relationCounts, error) {
	outVal, _ := rec.Get("outgoing")
	inVal, _ := rec.Get("incoming")
	return &relationCounts{Outgoing: toInt64(outVal), Incoming: toInt64(inVal)}, nil
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case int32:
		return int64(n)
	default:
		return 0
	}
}

func (r *Neo4jExtensionRepository) DeleteByPlugin(ctx context.Context, pluginID string) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ExtensionRepository/DeleteByPlugin")
	defer span.End()

	prefix := elemoplugin.RelationPrefix(pluginID)
	cypher := `
	MATCH ()-[r]->()
	WHERE type(r) STARTS WITH $prefix
	DELETE r
	WITH count(*) AS _
	MATCH (e:` + model.ResourceTypeExtension.String() + ` {plugin_id: $plugin_id})
	DETACH DELETE e`
	err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, map[string]any{
		"prefix":    prefix,
		"plugin_id": pluginID,
	})
	if err != nil {
		return errors.Join(ErrExtensionDelete, err)
	}
	return nil
}
