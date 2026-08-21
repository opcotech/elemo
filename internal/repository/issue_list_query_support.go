package repository

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

type IssueListSortField string

const (
	IssueListSortFieldRank      IssueListSortField = "rank"
	IssueListSortFieldTitle     IssueListSortField = "title"
	IssueListSortFieldPriority  IssueListSortField = "priority"
	IssueListSortFieldStatus    IssueListSortField = "status"
	IssueListSortFieldDueDate   IssueListSortField = "due_date"
	IssueListSortFieldCreatedAt IssueListSortField = "created_at"
	IssueListSortFieldUpdatedAt IssueListSortField = "updated_at"
	IssueListSortFieldID        IssueListSortField = "id"
)

type IssueListSort struct {
	Field     IssueListSortField
	Direction SortDirection
}

func (s IssueListSort) normalize() IssueListSort {
	field := s.Field
	switch field {
	case IssueListSortFieldRank,
		IssueListSortFieldTitle,
		IssueListSortFieldPriority,
		IssueListSortFieldStatus,
		IssueListSortFieldDueDate,
		IssueListSortFieldCreatedAt,
		IssueListSortFieldUpdatedAt,
		IssueListSortFieldID:
	default:
		field = IssueListSortFieldRank
	}

	direction := s.Direction
	if !direction.Valid() {
		direction = SortDirectionAsc
	}

	return IssueListSort{
		Field:     field,
		Direction: direction,
	}
}

func (s IssueListSort) nullable() bool {
	switch s.Field {
	case IssueListSortFieldDueDate, IssueListSortFieldUpdatedAt:
		return true
	default:
		return false
	}
}

func issuePriorityRank(priority model.IssuePriority) int64 {
	switch priority {
	case model.IssuePriorityLowest:
		return 0
	case model.IssuePriorityLow:
		return 1
	case model.IssuePriorityNormal:
		return 2
	case model.IssuePriorityHigh:
		return 3
	case model.IssuePriorityHighest:
		return 4
	default:
		return 99
	}
}

func issueStatusRank(status model.IssueStatus) int64 {
	switch status {
	case model.IssueStatusOpen:
		return 0
	case model.IssueStatusInProgress:
		return 1
	case model.IssueStatusReview:
		return 2
	case model.IssueStatusBlocked:
		return 3
	case model.IssueStatusDone:
		return 4
	case model.IssueStatusClosed:
		return 5
	default:
		return 99
	}
}

func issueListSortExpression(alias string, field IssueListSortField) string {
	switch field {
	case IssueListSortFieldRank:
		return alias + ".numeric_id"
	case IssueListSortFieldTitle:
		return "toLower(" + alias + ".title)"
	case IssueListSortFieldPriority:
		return "CASE " + alias + ".priority " +
			"WHEN 'lowest' THEN 0 " +
			"WHEN 'low' THEN 1 " +
			"WHEN 'normal' THEN 2 " +
			"WHEN 'high' THEN 3 " +
			"WHEN 'highest' THEN 4 " +
			"ELSE 99 END"
	case IssueListSortFieldStatus:
		return "CASE " + alias + ".status " +
			"WHEN 'open' THEN 0 " +
			"WHEN 'in progress' THEN 1 " +
			"WHEN 'review' THEN 2 " +
			"WHEN 'blocked' THEN 3 " +
			"WHEN 'done' THEN 4 " +
			"WHEN 'closed' THEN 5 " +
			"ELSE 99 END"
	case IssueListSortFieldDueDate:
		return alias + ".due_date"
	case IssueListSortFieldCreatedAt:
		return alias + ".created_at"
	case IssueListSortFieldUpdatedAt:
		return alias + ".updated_at"
	case IssueListSortFieldID:
		return alias + ".id"
	default:
		return alias + ".numeric_id"
	}
}

func issueListOrderClause(alias string, sort IssueListSort) string {
	expr := issueListSortExpression(alias, sort.Field)
	orderDir := sort.Direction.Cypher()
	tieDir := sort.Direction.Cypher()
	if sort.nullable() {
		return "ORDER BY CASE WHEN " + expr + " IS NULL THEN 1 ELSE 0 END ASC, " + expr + " " + orderDir + ", " + alias + ".id " + tieDir
	}

	return "ORDER BY " + expr + " " + orderDir + ", " + alias + ".id " + tieDir
}

type IssueListFilter struct {
	Text       string
	Statuses   []model.IssueStatus
	Priorities []model.IssuePriority
}

func normalizeIssueListFilter(filter IssueListFilter) IssueListFilter {
	out := IssueListFilter{
		Text: strings.ToLower(strings.TrimSpace(filter.Text)),
	}

	statusSeen := make(map[model.IssueStatus]struct{}, len(filter.Statuses))
	for _, status := range filter.Statuses {
		if !status.IsAIssueStatus() {
			continue
		}
		if _, ok := statusSeen[status]; ok {
			continue
		}
		statusSeen[status] = struct{}{}
		out.Statuses = append(out.Statuses, status)
	}

	prioritySeen := make(map[model.IssuePriority]struct{}, len(filter.Priorities))
	for _, priority := range filter.Priorities {
		if !priority.IsAIssuePriority() {
			continue
		}
		if _, ok := prioritySeen[priority]; ok {
			continue
		}
		prioritySeen[priority] = struct{}{}
		out.Priorities = append(out.Priorities, priority)
	}

	for i := 1; i < len(out.Statuses); i++ {
		j := i
		for j > 0 && out.Statuses[j-1].String() > out.Statuses[j].String() {
			out.Statuses[j-1], out.Statuses[j] = out.Statuses[j], out.Statuses[j-1]
			j--
		}
	}
	for i := 1; i < len(out.Priorities); i++ {
		j := i
		for j > 0 && out.Priorities[j-1].String() > out.Priorities[j].String() {
			out.Priorities[j-1], out.Priorities[j] = out.Priorities[j], out.Priorities[j-1]
			j--
		}
	}

	return out
}

func issueListFilterWhere(issueAlias, projectAlias string, filter IssueListFilter, params map[string]any) string {
	parts := make([]string, 0, 3)

	if filter.Text != "" {
		params["q"] = filter.Text
		parts = append(parts, "(toLower("+issueAlias+".title) CONTAINS $q OR toLower(coalesce("+issueAlias+".description, '')) CONTAINS $q OR toLower(coalesce("+projectAlias+".key, '') + '-' + toString("+issueAlias+".numeric_id)) CONTAINS $q)")
	}
	if len(filter.Statuses) > 0 {
		statuses := make([]string, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			statuses = append(statuses, status.String())
		}
		params["statuses"] = statuses
		parts = append(parts, issueAlias+".status IN $statuses")
	}
	if len(filter.Priorities) > 0 {
		priorities := make([]string, 0, len(filter.Priorities))
		for _, priority := range filter.Priorities {
			priorities = append(priorities, priority.String())
		}
		params["priorities"] = priorities
		parts = append(parts, issueAlias+".priority IN $priorities")
	}

	return strings.Join(parts, " AND ")
}

type issueListCursorPayload struct {
	Version   int                `json:"v"`
	Hash      string             `json:"h"`
	Field     IssueListSortField `json:"f"`
	Direction string             `json:"d"`
	ID        string             `json:"id"`
	Type      string             `json:"type"`
	Sort      *string            `json:"s,omitempty"`
	SortNull  bool               `json:"sn,omitempty"`
}

type issueListCursor struct {
	ID       model.ID
	Sort     *string
	SortNull bool
}

func issueListCursorHash(scopeID model.ID, filter IssueListFilter, sort IssueListSort) string {
	type hashInput struct {
		Scope      string   `json:"scope"`
		SortField  string   `json:"sort_field"`
		Direction  string   `json:"direction"`
		Text       string   `json:"text"`
		Statuses   []string `json:"statuses"`
		Priorities []string `json:"priorities"`
	}

	statuses := make([]string, 0, len(filter.Statuses))
	for _, status := range filter.Statuses {
		statuses = append(statuses, status.String())
	}

	priorities := make([]string, 0, len(filter.Priorities))
	for _, priority := range filter.Priorities {
		priorities = append(priorities, priority.String())
	}

	raw, _ := json.Marshal(hashInput{
		Scope:      scopeID.String(),
		SortField:  string(sort.Field),
		Direction:  sort.Direction.String(),
		Text:       filter.Text,
		Statuses:   statuses,
		Priorities: priorities,
	})

	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:16])
}

func encodeIssueListCursor(issue *PartialIssue, sort IssueListSort, cursorHash string) (string, error) {
	if issue == nil {
		return "", ErrInvalidCursor
	}

	sortValue, sortNull, err := issueListSortValue(issue, sort.Field)
	if err != nil {
		return "", err
	}

	payload := issueListCursorPayload{
		Version:   1,
		Hash:      cursorHash,
		Field:     sort.Field,
		Direction: sort.Direction.String(),
		ID:        issue.ID.String(),
		Type:      issue.ID.Label(),
		Sort:      sortValue,
		SortNull:  sortNull,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", errors.Join(ErrInvalidCursor, err)
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeIssueListCursor(token string, cursorHash string, sort IssueListSort) (issueListCursor, error) {
	if token == "" {
		return issueListCursor{}, ErrInvalidCursor
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return issueListCursor{}, errors.Join(ErrInvalidCursor, err)
	}

	var payload issueListCursorPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return issueListCursor{}, errors.Join(ErrInvalidCursor, err)
	}
	if payload.Version != 1 ||
		payload.Hash == "" ||
		payload.Field == "" ||
		payload.Direction == "" ||
		payload.ID == "" ||
		payload.Type == "" {
		return issueListCursor{}, ErrInvalidCursor
	}
	direction, err := SortDirectionString(payload.Direction)
	if err != nil || !direction.Valid() {
		return issueListCursor{}, ErrInvalidCursor
	}
	if payload.Hash != cursorHash ||
		payload.Field != sort.Field ||
		direction != sort.Direction {
		return issueListCursor{}, ErrInvalidCursor
	}

	id, err := model.NewIDFromString(payload.ID, payload.Type)
	if err != nil {
		return issueListCursor{}, errors.Join(ErrInvalidCursor, err)
	}
	if id.Type != model.ResourceTypeIssue {
		return issueListCursor{}, ErrInvalidCursor
	}

	if payload.Sort == nil && !payload.SortNull && sort.Field != IssueListSortFieldID {
		return issueListCursor{}, ErrInvalidCursor
	}

	return issueListCursor{
		ID:       id,
		Sort:     payload.Sort,
		SortNull: payload.SortNull,
	}, nil
}

func issueListSortValue(issue *PartialIssue, field IssueListSortField) (*string, bool, error) {
	switch field {
	case IssueListSortFieldRank:
		v := strconv.FormatUint(uint64(issue.NumericID), 10)
		return &v, false, nil
	case IssueListSortFieldTitle:
		v := strings.ToLower(issue.Title)
		return &v, false, nil
	case IssueListSortFieldPriority:
		v := strconv.FormatInt(issuePriorityRank(issue.Priority), 10)
		return &v, false, nil
	case IssueListSortFieldStatus:
		v := strconv.FormatInt(issueStatusRank(issue.Status), 10)
		return &v, false, nil
	case IssueListSortFieldDueDate:
		if issue.DueDate == nil {
			return nil, true, nil
		}
		v := issue.DueDate.UTC().Format(timeFormatNano)
		return &v, false, nil
	case IssueListSortFieldCreatedAt:
		if issue.CreatedAt == nil {
			return nil, true, nil
		}
		v := issue.CreatedAt.UTC().Format(timeFormatNano)
		return &v, false, nil
	case IssueListSortFieldUpdatedAt:
		if issue.UpdatedAt == nil {
			return nil, true, nil
		}
		v := issue.UpdatedAt.UTC().Format(timeFormatNano)
		return &v, false, nil
	case IssueListSortFieldID:
		v := issue.ID.String()
		return &v, false, nil
	default:
		return nil, false, ErrUnsupportedOrder
	}
}

const timeFormatNano = "2006-01-02T15:04:05.999999999Z07:00"

func issueListCursorWhere(alias string, sort IssueListSort, cursor issueListCursor, params map[string]any) (string, error) {
	sort = sort.normalize()
	params["cursor_id"] = cursor.ID.String()
	if sort.Field == IssueListSortFieldID {
		return CursorWhereCypher(alias, sort.Direction)
	}

	expr := issueListSortExpression(alias, sort.Field)
	cursorValue := "cursor_sort"
	if !cursor.SortNull {
		if cursor.Sort == nil {
			return "", ErrInvalidCursor
		}
		switch sort.Field {
		case IssueListSortFieldRank, IssueListSortFieldPriority, IssueListSortFieldStatus:
			n, err := strconv.ParseInt(*cursor.Sort, 10, 64)
			if err != nil {
				return "", errors.Join(ErrInvalidCursor, err)
			}
			params[cursorValue] = n
		default:
			params[cursorValue] = *cursor.Sort
		}
	}

	greaterThan := ">"
	if sort.Direction == SortDirectionDesc {
		greaterThan = "<"
	}

	if sort.nullable() {
		if cursor.SortNull {
			return fmt.Sprintf("(%s IS NULL AND %s.id %s $cursor_id)", expr, alias, greaterThan), nil
		}
		return fmt.Sprintf("(%s IS NULL OR %s %s $%s OR (%s = $%s AND %s.id %s $cursor_id))", expr, expr, greaterThan, cursorValue, expr, cursorValue, alias, greaterThan), nil
	}

	return fmt.Sprintf("(%s %s $%s OR (%s = $%s AND %s.id %s $cursor_id))", expr, greaterThan, cursorValue, expr, cursorValue, alias, greaterThan), nil
}
