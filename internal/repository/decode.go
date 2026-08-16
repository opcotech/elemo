package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j/dbtype"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

// Neo4jNodeProperty returns a typed property from a node.
func Neo4jNodeProperty[T neo4j.PropertyValue](node neo4j.Node, key string) (T, error) {
	return neo4j.GetProperty[T](node, key)
}

// Neo4jOptionalNodeProperty returns a typed property or zero when missing.
func Neo4jOptionalNodeProperty[T neo4j.PropertyValue](node neo4j.Node, key string) (T, bool, error) {
	var zero T
	if _, ok := node.Props[key]; !ok {
		return zero, false, nil
	}
	val, err := neo4j.GetProperty[T](node, key)
	if err != nil {
		return zero, false, err
	}
	return val, true, nil
}

// Neo4jDecodeID reads an ID property and builds a model.ID with the given type.
func Neo4jDecodeID(node neo4j.Node, resourceType model.ResourceType) (model.ID, error) {
	idStr, err := Neo4jNodeProperty[string](node, "id")
	if err != nil {
		return model.ID{}, errors.Join(ErrMalformedResult, err)
	}
	id, err := model.NewIDFromString(idStr, resourceType.String())
	if err != nil {
		return model.ID{}, errors.Join(ErrMalformedResult, err)
	}
	return id, nil
}

// Neo4jDecodeIDFromLabel reads an ID using the first node label as resource type.
func Neo4jDecodeIDFromLabel(node neo4j.Node) (model.ID, error) {
	if len(node.Labels) == 0 {
		return model.ID{}, ErrMalformedResult
	}
	idStr, err := Neo4jNodeProperty[string](node, "id")
	if err != nil {
		return model.ID{}, errors.Join(ErrMalformedResult, err)
	}
	id, err := model.NewIDFromString(idStr, node.Labels[0])
	if err != nil {
		return model.ID{}, errors.Join(ErrMalformedResult, err)
	}
	return id, nil
}

// Neo4jDecodeTime parses a temporal property into *time.Time.
func Neo4jDecodeTime(val any) (*time.Time, error) {
	if val == nil {
		return nil, nil
	}
	switch t := val.(type) {
	case time.Time:
		tt := t.UTC()
		return &tt, nil
	case dbtype.LocalDateTime:
		tt := time.Time(t).UTC()
		return &tt, nil
	case dbtype.Date:
		tt := time.Time(t).UTC()
		return &tt, nil
	case string:
		parsed, err := time.Parse(time.RFC3339Nano, t)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, t)
			if err != nil {
				return nil, errors.Join(ErrMalformedResult, err)
			}
		}
		parsed = parsed.UTC()
		return &parsed, nil
	default:
		return nil, errors.Join(ErrMalformedResult, fmt.Errorf("unsupported time type %T", val))
	}
}

// Neo4jNodeTime reads a temporal property from a node.
func Neo4jNodeTime(node neo4j.Node, key string) (*time.Time, error) {
	val, ok := node.Props[key]
	if !ok || val == nil {
		return nil, nil
	}
	return Neo4jDecodeTime(val)
}

// Neo4jRecordNode returns a node by key from a record.
func Neo4jRecordNode(record *neo4j.Record, key string) (neo4j.Node, error) {
	node, _, err := neo4j.GetRecordValue[neo4j.Node](record, key)
	if err != nil {
		return neo4j.Node{}, errors.Join(ErrMalformedResult, err)
	}
	return node, nil
}

// Neo4jRecordOptionalNode returns a node by key, or nil when the value is absent.
func Neo4jRecordOptionalNode(record *neo4j.Record, key string) (*neo4j.Node, error) {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return nil, nil
	}
	node, ok := val.(neo4j.Node)
	if !ok {
		return nil, ErrMalformedResult
	}
	return &node, nil
}

func mapString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// Neo4jRecordIDs decodes a list of string IDs from a record into model.IDs.
func Neo4jRecordIDs(record *neo4j.Record, key string, resourceType model.ResourceType) ([]model.ID, error) {
	val, err := Neo4jParseValueFromRecord[[]any](record, key)
	if err != nil {
		return nil, err
	}
	ids := make([]model.ID, 0, len(val))
	for _, item := range val {
		if item == nil {
			continue
		}
		idStr, ok := item.(string)
		if !ok {
			return nil, ErrMalformedResult
		}
		id, err := model.NewIDFromString(idStr, resourceType.String())
		if err != nil {
			return nil, errors.Join(ErrMalformedResult, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Neo4jScanNodeScalars copies scalar node properties into dst using the legacy
// JSON path for types not yet migrated to explicit decoders.
func Neo4jScanNodeScalars(node neo4j.Node, dst any, exclude []string) error {
	return Neo4jScanIntoStruct(&node, dst, exclude)
}

// Neo4jScanIntoStruct parses a struct from a neo4j node or relationship.
func Neo4jScanIntoStruct(n Neo4jPropertyGetter, dst any, exclude []string) error {
	props := make(map[string]any)

	for k, v := range n.GetProperties() {
		props[k] = v
	}

	for _, e := range exclude {
		delete(props, e)
	}

	return convert.AnyToAny(props, dst)
}

// Neo4jParseValueFromRecord parses a value from a neo4j record.
func Neo4jParseValueFromRecord[T neo4j.RecordValue](record *neo4j.Record, key string) (T, error) {
	var zero T

	value, _, err := neo4j.GetRecordValue[T](record, key)
	if err != nil {
		return zero, errors.Join(ErrMalformedResult, err)
	}

	return value, nil
}

// Neo4jParseIDsFromRecord parses a list of IDs from a neo4j record.
func Neo4jParseIDsFromRecord(record *neo4j.Record, key, label string) ([]model.ID, error) {
	val, err := Neo4jParseValueFromRecord[[]any](record, key)
	if err != nil {
		return nil, err
	}

	ids := make([]model.ID, len(val))
	for i, p := range val {
		id, err := model.NewIDFromString(p.(string), label)
		if err != nil {
			return nil, err
		}

		ids[i] = id
	}

	return ids, nil
}

func partialProjectFromNode(node neo4j.Node) (*PartialProject, error) {
	projectID, err := Neo4jDecodeID(node, model.ResourceTypeProject)
	if err != nil {
		return nil, err
	}

	var tempProject struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Logo        string `json:"logo"`
		Status      string `json:"status"`
	}
	if err := Neo4jScanIntoStruct(&node, &tempProject, []string{"id"}); err != nil {
		return nil, err
	}

	var status model.ProjectStatus
	if err := status.UnmarshalText([]byte(tempProject.Status)); err != nil {
		return nil, err
	}

	return &PartialProject{
		ID:          projectID,
		Key:         tempProject.Key,
		Name:        tempProject.Name,
		Description: tempProject.Description,
		Logo:        tempProject.Logo,
		Status:      status,
	}, nil
}

func Neo4jRecordPartialProject(record *neo4j.Record, key string) (*PartialProject, error) {
	node, err := Neo4jRecordNode(record, key)
	if err != nil {
		return nil, err
	}
	return partialProjectFromNode(node)
}

func partialNamespaceFromNode(node neo4j.Node) (*PartialNamespace, error) {
	id, err := Neo4jDecodeID(node, model.ResourceTypeNamespace)
	if err != nil {
		return nil, err
	}
	name, err := Neo4jNodeProperty[string](node, "name")
	if err != nil {
		return nil, errors.Join(ErrMalformedResult, err)
	}
	return &PartialNamespace{ID: id, Name: name}, nil
}

func Neo4jRecordOptionalPartialNamespace(record *neo4j.Record, key string) (*PartialNamespace, error) {
	node, err := Neo4jRecordOptionalNode(record, key)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}
	return partialNamespaceFromNode(*node)
}

func partialLabelFromNode(node neo4j.Node) (PartialLabel, error) {
	id, err := Neo4jDecodeID(node, model.ResourceTypeLabel)
	if err != nil {
		return PartialLabel{}, err
	}
	name, err := Neo4jNodeProperty[string](node, "name")
	if err != nil {
		return PartialLabel{}, errors.Join(ErrMalformedResult, err)
	}
	return PartialLabel{ID: id, Name: name}, nil
}

func parsePartialLabels(val any) ([]PartialLabel, error) {
	if val == nil {
		return make([]PartialLabel, 0), nil
	}
	items, ok := val.([]any)
	if !ok {
		return nil, ErrMalformedResult
	}
	labels := make([]PartialLabel, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		node, ok := item.(neo4j.Node)
		if !ok {
			return nil, ErrMalformedResult
		}
		label, err := partialLabelFromNode(node)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	return labels, nil
}

func Neo4jRecordPartialLabels(record *neo4j.Record, key string) ([]PartialLabel, error) {
	val, err := Neo4jParseValueFromRecord[[]any](record, key)
	if err != nil {
		return nil, err
	}
	return parsePartialLabels(val)
}

func partialUserFromNode(node neo4j.Node) (*PartialUser, error) {
	id, err := Neo4jDecodeID(node, model.ResourceTypeUser)
	if err != nil {
		return nil, err
	}
	firstName, err := Neo4jNodeProperty[string](node, "first_name")
	if err != nil {
		return nil, errors.Join(ErrMalformedResult, err)
	}
	lastName, err := Neo4jNodeProperty[string](node, "last_name")
	if err != nil {
		return nil, errors.Join(ErrMalformedResult, err)
	}
	picture, _, err := Neo4jOptionalNodeProperty[string](node, "picture")
	if err != nil {
		return nil, errors.Join(ErrMalformedResult, err)
	}
	return &PartialUser{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Picture:   picture,
	}, nil
}

func Neo4jRecordPartialUser(record *neo4j.Record, key string) (*PartialUser, error) {
	node, err := Neo4jRecordOptionalNode(record, key)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}
	return partialUserFromNode(*node)
}
