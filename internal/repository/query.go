package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

// CompiledQuery is a named, parameterized Cypher statement produced by the
// internal query compiler. Cypher structure is allowlisted; callers never pass
// raw fragments from HTTP.
type CompiledQuery struct {
	Name   string
	Cypher string
	Params map[string]any
}

// Fingerprint returns a stable hash of the query name and Cypher text for
// cache keys and observability. Parameters are excluded.
func (q CompiledQuery) Fingerprint() string {
	sum := sha256.Sum256([]byte(q.Name + "\n" + q.Cypher))
	return hex.EncodeToString(sum[:16])
}

// CacheKey returns a cache key composed from the query fingerprint and params.
func (q CompiledQuery) CacheKey(prefix string, extra ...any) string {
	parts := make([]any, 0, 2+len(extra)+len(q.Params))
	parts = append(parts, prefix, q.Fingerprint())
	parts = append(parts, extra...)
	for _, k := range sortedParamKeys(q.Params) {
		parts = append(parts, k, fmt.Sprintf("%v", q.Params[k]))
	}
	return composeCacheKey(parts...)
}

func sortedParamKeys(params map[string]any) []string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	// insertion sort keeps the file dependency-free
	for i := 1; i < len(keys); i++ {
		j := i
		for j > 0 && keys[j-1] > keys[j] {
			keys[j-1], keys[j] = keys[j], keys[j-1]
			j--
		}
	}
	return keys
}

// QueryCompiler produces a QueryPlan from an allowlisted query struct.
type QueryCompiler interface {
	Compile() (QueryPlan, error)
}

// CompileQuery compiles q into a QueryPlan.
func CompileQuery(q QueryCompiler) (QueryPlan, error) {
	return q.Compile()
}

// QueryPlan is a root query plus optional relation loaders executed in one
// managed read transaction.
type QueryPlan struct {
	Root    CompiledQuery
	Loaders []CompiledQuery
}

// Fingerprint returns a stable hash of the full plan shape, including loader
// Cypher, so projection-specific loaders participate in cache keys.
func (p QueryPlan) Fingerprint() string {
	var builder strings.Builder
	builder.WriteString(p.Root.Name)
	builder.WriteString("\n")
	builder.WriteString(p.Root.Cypher)
	for _, loader := range p.Loaders {
		builder.WriteString("\n---\n")
		builder.WriteString(loader.Name)
		builder.WriteString("\n")
		builder.WriteString(loader.Cypher)
	}

	sum := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(sum[:16])
}

// CacheKey returns a cache key composed from the full plan fingerprint and root
// params. Loader params are structural here and already covered by Fingerprint.
func (p QueryPlan) CacheKey(prefix string, extra ...any) string {
	parts := make([]any, 0, 2+len(extra)+len(p.Root.Params))
	parts = append(parts, prefix, p.Fingerprint())
	parts = append(parts, extra...)
	for _, key := range sortedParamKeys(p.Root.Params) {
		parts = append(parts, key, fmt.Sprintf("%v", p.Root.Params[key]))
	}
	return composeCacheKey(parts...)
}

// Validate ensures the plan has a named root query and named loaders.
func (p QueryPlan) Validate() error {
	if strings.TrimSpace(p.Root.Name) == "" || strings.TrimSpace(p.Root.Cypher) == "" {
		return ErrQueryCompile
	}
	for _, loader := range p.Loaders {
		if strings.TrimSpace(loader.Name) == "" || strings.TrimSpace(loader.Cypher) == "" {
			return ErrQueryCompile
		}
	}
	return nil
}

// CursorWhereCypher returns an allowlisted WHERE fragment for id continuation.
// direction must be ASC or DESC. XIDs are k-sortable, so id alone is a fully
// unique order; stored created_at is not recovered from the id.
func CursorWhereCypher(nodeAlias string, direction SortDirection) (string, error) {
	if !direction.Valid() {
		return "", ErrUnsupportedOrder
	}
	if direction == SortDirectionDesc {
		return fmt.Sprintf("%s.id < $cursor_id", nodeAlias), nil
	}
	return fmt.Sprintf("%s.id > $cursor_id", nodeAlias), nil
}

// ApplyCursorParams mutates params with cursor bound values when present.
func ApplyCursorParams(params map[string]any, token *string) error {
	if token == nil || *token == "" {
		return nil
	}
	id, err := DecodeCursor(*token)
	if err != nil {
		return err
	}
	params["cursor_id"] = id.String()
	return nil
}

// cursorBounds is the shared result of compiling list pagination: normalized
// page, allowlisted order, and an optional raw cursor predicate (no WHERE/AND).
type cursorBounds struct {
	Page  CursorPage
	Order SortDirection
	Where string
}

// compileCursorBounds normalizes page, defaults order to DESC, sets $limit, and
// when a cursor token is present fills $cursor_* params. Empty tokens are
// already coerced to nil by CursorPage.Normalize, so no WHERE is emitted.
func compileCursorBounds(alias string, page CursorPage, order SortDirection, params map[string]any) (cursorBounds, error) {
	normalized, err := page.Normalize()
	if err != nil {
		return cursorBounds{}, err
	}
	if order == SortDirectionUnknown {
		order = SortDirectionDesc
	}
	if !order.Valid() {
		return cursorBounds{}, ErrUnsupportedOrder
	}

	params["limit"] = normalized.FetchLimit()

	var where string
	if normalized.Token != nil {
		cursorWhere, err := CursorWhereCypher(alias, order)
		if err != nil {
			return cursorBounds{}, err
		}
		if err := ApplyCursorParams(params, normalized.Token); err != nil {
			return cursorBounds{}, err
		}
		where = cursorWhere
	}

	return cursorBounds{Page: normalized, Order: order, Where: where}, nil
}

func cursorWherePrefix(where, prefix string) string {
	if where == "" {
		return ""
	}
	return prefix + where
}

func whereClause(prefix string, parts ...string) string {
	joined := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			joined = append(joined, part)
		}
	}
	if len(joined) == 0 {
		return ""
	}
	return prefix + strings.Join(joined, " AND ")
}

func cloneParams(params map[string]any) map[string]any {
	cloned := make(map[string]any, len(params))
	for key, value := range params {
		cloned[key] = value
	}
	return cloned
}

// QuerySummary captures observability fields from a Neo4j result summary.
type QuerySummary struct {
	Name                 string
	Fingerprint          string
	Counters             neo4j.Counters
	ResultAvailableAfter time.Duration
	ResultConsumedAfter  time.Duration
}
