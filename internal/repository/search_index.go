package repository

import (
	"context"
	"errors"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
)

// SearchableRecord is the graph projection used to build a SearchDocument.
// Ancestry is the resource plus IN_SCOPE_OF ancestors, nearest first.
type SearchableRecord struct {
	ID        model.ID
	Title     string
	Content   string
	Key       string
	CreatedAt int64
	UpdatedAt int64
	Ancestry  []model.ID
}

func searchableRecordAncestryCypher() string {
	return `
	COLLECT {
		MATCH path = (n)-[:` + EdgeKindInScopeOf.String() + `*0..]->(scope)
		WHERE ` + authzAcyclicPathPredicate("path") + `
		WITH DISTINCT scope, min(length(path)) AS depth
		ORDER BY depth
		RETURN scope
	} AS ancestry`
}

func scanSearchableRecord(resourceType model.ResourceType) func(rec *neo4j.Record) (SearchableRecord, error) {
	return func(rec *neo4j.Record) (SearchableRecord, error) {
		node, err := Neo4jRecordNode(rec, "n")
		if err != nil {
			return SearchableRecord{}, err
		}
		ancestryNodes, err := neo4jRecordNodes(rec, "ancestry")
		if err != nil {
			return SearchableRecord{}, err
		}
		return decodeSearchableRecord(resourceType, node, ancestryNodes)
	}
}

func neo4jRecordNodes(rec *neo4j.Record, key string) ([]neo4j.Node, error) {
	val, ok := rec.Get(key)
	if !ok || val == nil {
		return []neo4j.Node{}, nil
	}
	switch items := val.(type) {
	case []neo4j.Node:
		return items, nil
	case []any:
		nodes := make([]neo4j.Node, 0, len(items))
		for _, item := range items {
			if item == nil {
				continue
			}
			node, ok := item.(neo4j.Node)
			if !ok {
				return nil, ErrMalformedResult
			}
			nodes = append(nodes, node)
		}
		return nodes, nil
	default:
		return nil, ErrMalformedResult
	}
}

func optionalNodeString(node neo4j.Node, key string) string {
	val, ok, err := Neo4jOptionalNodeProperty[string](node, key)
	if err != nil || !ok {
		return ""
	}
	return val
}

func unixSecondsFromNode(node neo4j.Node, key string) (int64, error) {
	t, err := Neo4jNodeTime(node, key)
	if err != nil {
		return 0, err
	}
	if t == nil || t.IsZero() {
		return 0, nil
	}
	return t.Unix(), nil
}

func decodeSearchableRecord(resourceType model.ResourceType, node neo4j.Node, ancestryNodes []neo4j.Node) (SearchableRecord, error) {
	id, err := Neo4jDecodeID(node, resourceType)
	if err != nil {
		return SearchableRecord{}, err
	}

	ancestry := make([]model.ID, 0, len(ancestryNodes))
	projectKey := ""
	for _, ancestor := range ancestryNodes {
		scope, decErr := Neo4jDecodeIDFromLabel(ancestor)
		if decErr != nil {
			return SearchableRecord{}, decErr
		}
		ancestry = append(ancestry, scope)
		if scope.Type == model.ResourceTypeProject {
			projectKey = optionalNodeString(ancestor, "key")
		}
	}
	if len(ancestry) == 0 {
		ancestry = []model.ID{id}
	}

	createdAt, err := unixSecondsFromNode(node, "created_at")
	if err != nil {
		return SearchableRecord{}, err
	}
	updatedAt, err := unixSecondsFromNode(node, "updated_at")
	if err != nil {
		return SearchableRecord{}, err
	}

	record := SearchableRecord{
		ID:        id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Ancestry:  ancestry,
	}

	switch resourceType {
	case model.ResourceTypeOrganization:
		record.Title = optionalNodeString(node, "name")
	case model.ResourceTypeNamespace:
		record.Title = optionalNodeString(node, "name")
		record.Content = optionalNodeString(node, "description")
	case model.ResourceTypeProject:
		record.Title = optionalNodeString(node, "name")
		record.Content = optionalNodeString(node, "description")
		record.Key = optionalNodeString(node, "key")
	case model.ResourceTypeIssue:
		record.Title = optionalNodeString(node, "title")
		record.Content = optionalNodeString(node, "description")
		numericID, ok, numErr := Neo4jOptionalNodeProperty[int64](node, "numeric_id")
		if numErr != nil {
			return SearchableRecord{}, errors.Join(ErrMalformedResult, numErr)
		}
		if ok && projectKey != "" {
			record.Key = model.FormatIssueKey(projectKey, uint(numericID)) // nolint:gosec
		}
	case model.ResourceTypeDocument:
		record.Title = optionalNodeString(node, "title")
		record.Content = optionalNodeString(node, "excerpt")
	default:
		return SearchableRecord{}, model.ErrInvalidResourceType
	}

	return record, nil
}

func validateSearchIndexRead(db *Neo4jDatabase, resourceType model.ResourceType) error {
	if db == nil {
		return ErrNoDriver
	}
	if !resourceType.IsAResourceType() {
		return model.ErrInvalidResourceType
	}
	return nil
}

// ListSearchableIDs returns every node ID of resourceType. Used by reindex.
func ListSearchableIDs(ctx context.Context, db *Neo4jDatabase, resourceType model.ResourceType) ([]model.ID, error) {
	if err := validateSearchIndexRead(db, resourceType); err != nil {
		return nil, err
	}

	cypher := `MATCH (n:` + resourceType.String() + `) RETURN n.id AS id ORDER BY n.id`
	ids, err := Neo4jExecuteReadAndReadAll(ctx, db, cypher, nil, func(rec *neo4j.Record) (model.ID, error) {
		raw, err := Neo4jParseValueFromRecord[string](rec, "id")
		if err != nil {
			return model.ID{}, err
		}
		return model.NewIDFromString(raw, resourceType.String())
	})
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []model.ID{}
	}
	return ids, nil
}

// ListSearchableRecords returns a page of searchable graph rows for
// resourceType, ordered by node id. An empty after starts at the beginning.
func ListSearchableRecords(
	ctx context.Context,
	db *Neo4jDatabase,
	resourceType model.ResourceType,
	after string,
	limit int,
) ([]SearchableRecord, error) {
	if err := validateSearchIndexRead(db, resourceType); err != nil {
		return nil, err
	}
	if limit < MinPageSize || limit > MaxPageSize {
		return nil, ErrInvalidPageSize
	}

	cypher := `
	MATCH (n:` + resourceType.String() + `)
	WHERE $after = "" OR n.id > $after
	WITH n
	ORDER BY n.id
	LIMIT $limit
	RETURN n, ` + searchableRecordAncestryCypher()

	records, err := Neo4jExecuteReadAndReadAll(
		ctx,
		db,
		cypher,
		map[string]any{
			"after": after,
			"limit": limit,
		},
		scanSearchableRecord(resourceType),
	)
	if err != nil {
		return nil, err
	}

	if records == nil {
		return []SearchableRecord{}, nil
	}

	return records, nil
}

// ListSearchableRecordsByIDs returns searchable rows for the given IDs of
// resourceType. Missing nodes are omitted.
func ListSearchableRecordsByIDs(
	ctx context.Context,
	db *Neo4jDatabase,
	resourceType model.ResourceType,
	ids []model.ID,
) ([]SearchableRecord, error) {
	if err := validateSearchIndexRead(db, resourceType); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []SearchableRecord{}, nil
	}

	rawIDs := make([]string, len(ids))
	for i, id := range ids {
		rawIDs[i] = id.String()
	}

	cypher := `
	UNWIND $ids AS rid
	MATCH (n:` + resourceType.String() + ` {id: rid})
	RETURN n, ` + searchableRecordAncestryCypher()

	records, err := Neo4jExecuteReadAndReadAll(
		ctx,
		db,
		cypher,
		map[string]any{"ids": rawIDs},
		scanSearchableRecord(resourceType),
	)
	if err != nil {
		return nil, err
	}

	if records == nil {
		return []SearchableRecord{}, nil
	}

	return records, nil
}
