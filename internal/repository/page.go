package repository

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/opcotech/elemo/internal/model"
)

const (
	DefaultPageSize = 100
	MinPageSize     = 1
	MaxPageSize     = 1000
)

const (
	SortDirectionUnknown SortDirection = iota // SortDirectionUnknown
	SortDirectionAsc                          // ASC
	SortDirectionDesc                         // DESC
)

// SortDirection is an allowlisted order direction for compiled queries.
//
//go:generate go tool enumer -type=SortDirection -transform=noop -linecomment -output=page_sort_direction_gen.go
type SortDirection uint8

func (d SortDirection) Valid() bool {
	return d == SortDirectionAsc || d == SortDirectionDesc
}

func (d SortDirection) Cypher() string {
	return d.String()
}

// CursorPage describes a cursor-paginated list request.
type CursorPage struct {
	Size  int
	Token *string
}

// Normalize applies defaults and bounds to the page request.
func (p CursorPage) Normalize() (CursorPage, error) {
	out := p
	if out.Size == 0 {
		out.Size = DefaultPageSize
	}
	if out.Size < MinPageSize || out.Size > MaxPageSize {
		return CursorPage{}, errors.Join(ErrInvalidPageSize, fmt.Errorf("size %d not in [%d, %d]", out.Size, MinPageSize, MaxPageSize))
	}
	if out.Token != nil && *out.Token == "" {
		out.Token = nil
	}
	return out, nil
}

// FetchLimit returns page size + 1 used to detect a following page.
func (p CursorPage) FetchLimit() int {
	return p.Size + 1
}

// PageInfo holds opaque continuation metadata for a page of results.
type PageInfo struct {
	NextPageToken *string `json:"next_page_token,omitempty"`
	HasMore       bool    `json:"has_more"`
	TotalCount    *int64  `json:"total_count,omitempty"`
}

// Page is a bounded page of items with continuation metadata.
type Page[T any] struct {
	Items    []T      `json:"items"`
	PageInfo PageInfo `json:"page_info"`
}

// EmptyPage returns an empty page with no continuation.
func EmptyPage[T any]() Page[T] {
	return Page[T]{
		Items: make([]T, 0),
		PageInfo: PageInfo{
			HasMore: false,
		},
	}
}

// cursorPayload is the opaque token payload for XID ordering.
type cursorPayload struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// EncodeCursor encodes a stable id cursor token. XIDs are k-sortable and unique,
// so the id is a fully unique order without a created_at tiebreaker.
func EncodeCursor(id model.ID) (string, error) {
	if err := id.Validate(); err != nil {
		return "", errors.Join(ErrInvalidCursor, err)
	}

	raw, err := json.Marshal(cursorPayload{
		ID:   id.String(),
		Type: id.Label(),
	})
	if err != nil {
		return "", errors.Join(ErrInvalidCursor, err)
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// DecodeCursor decodes an opaque id cursor token.
func DecodeCursor(token string) (model.ID, error) {
	if token == "" {
		return model.ID{}, ErrInvalidCursor
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return model.ID{}, errors.Join(ErrInvalidCursor, err)
	}

	var payload cursorPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return model.ID{}, errors.Join(ErrInvalidCursor, err)
	}
	if payload.ID == "" || payload.Type == "" {
		return model.ID{}, ErrInvalidCursor
	}

	id, err := model.NewIDFromString(payload.ID, payload.Type)
	if err != nil {
		return model.ID{}, errors.Join(ErrInvalidCursor, err)
	}

	return id, nil
}

// PaginateSlice turns a fetch of size+1 into a Page and optional next token.
// id extracts the cursor field from the last returned item.
func PaginateSlice[T any](items []T, pageSize int, id func(T) model.ID) (Page[T], error) {
	if pageSize < MinPageSize {
		return Page[T]{}, ErrInvalidPageSize
	}

	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}

	page := Page[T]{
		Items: items,
		PageInfo: PageInfo{
			HasMore: hasMore,
		},
	}
	if !hasMore || len(items) == 0 {
		return page, nil
	}

	token, err := EncodeCursor(id(items[len(items)-1]))
	if err != nil {
		return Page[T]{}, err
	}
	page.PageInfo.NextPageToken = &token
	return page, nil
}
