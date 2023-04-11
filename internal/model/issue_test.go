package model

import (
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueKind_String(t *testing.T) {
	tests := []struct {
		name string
		k    IssueKind
		want string
	}{
		{"Epic", IssueKindEpic, "epic"},
		{"Story", IssueKindStory, "story"},
		{"Task", IssueKindTask, "task"},
		{"Bug", IssueKindBug, "bug"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.k.String())
		})
	}
}

func TestIssueKind_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		k        IssueKind
		wantText []byte
		wantErr  error
	}{
		{"Epic", IssueKindEpic, []byte("epic"), nil},
		{"Story", IssueKindStory, []byte("story"), nil},
		{"Task", IssueKindTask, []byte("task"), nil},
		{"Bug", IssueKindBug, []byte("bug"), nil},
		{"kind high", IssueKind(100), nil, ErrInvalidIssueKind},
		{"kind low", IssueKind(0), nil, ErrInvalidIssueKind},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.k.MarshalText()
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantText, gotText)
			}
		})
	}
}

func TestIssueKind_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		k       IssueKind
		args    args
		wantErr error
	}{
		{"Epic", IssueKindEpic, args{[]byte("epic")}, nil},
		{"Story", IssueKindStory, args{[]byte("story")}, nil},
		{"Task", IssueKindTask, args{[]byte("task")}, nil},
		{"Bug", IssueKindBug, args{[]byte("bug")}, nil},
		{"kind high", IssueKind(100), args{[]byte("")}, ErrInvalidIssueKind},
		{"kind low", IssueKind(0), args{[]byte("")}, ErrInvalidIssueKind},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.k.UnmarshalText(tt.args.text)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestIssuePriority_String(t *testing.T) {
	tests := []struct {
		name string
		p    IssuePriority
		want string
	}{
		{"Critical", IssuePriorityCritical, "critical"},
		{"High", IssuePriorityHigh, "high"},
		{"Medium", IssuePriorityMedium, "medium"},
		{"Low", IssuePriorityLow, "low"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.p.String())
		})
	}
}

func TestIssuePriority_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		p        IssuePriority
		wantText []byte
		wantErr  error
	}{
		{"Critical", IssuePriorityCritical, []byte("critical"), nil},
		{"High", IssuePriorityHigh, []byte("high"), nil},
		{"Medium", IssuePriorityMedium, []byte("medium"), nil},
		{"Low", IssuePriorityLow, []byte("low"), nil},
		{"priority high", IssuePriority(100), nil, ErrInvalidIssuePriority},
		{"priority low", IssuePriority(0), nil, ErrInvalidIssuePriority},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.p.MarshalText()
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantText, gotText)
			}
		})
	}
}

func TestIssuePriority_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		p       IssuePriority
		args    args
		wantErr error
	}{
		{"Critical", IssuePriorityCritical, args{[]byte("high")}, nil},
		{"High", IssuePriorityHigh, args{[]byte("high")}, nil},
		{"Medium", IssuePriorityMedium, args{[]byte("medium")}, nil},
		{"Low", IssuePriorityLow, args{[]byte("low")}, nil},
		{"priority high", IssuePriority(100), args{[]byte("")}, ErrInvalidIssuePriority},
		{"priority low", IssuePriority(0), args{[]byte("")}, ErrInvalidIssuePriority},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.p.UnmarshalText(tt.args.text)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestIssueRelationKind_String(t *testing.T) {
	tests := []struct {
		name string
		r    IssueRelationKind
		want string
	}{
		{"blocked by", IssueRelationKindBlockedBy, "blocked by"},
		{"blocks", IssueRelationKindBlocks, "blocks"},
		{"depends on", IssueRelationKindDependsOn, "depends on"},
		{"duplicated by", IssueRelationKindDuplicatedBy, "duplicated by"},
		{"duplicates", IssueRelationKindDuplicates, "duplicates"},
		{"related to", IssueRelationKindRelatedTo, "related to"},
		{"subtask of", IssueRelationKindSubtaskOf, "subtask of"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.r.String())
		})
	}
}

func TestIssueRelationKind_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		r        IssueRelationKind
		wantText []byte
		wantErr  error
	}{
		{"blocked by", IssueRelationKindBlockedBy, []byte("blocked by"), nil},
		{"blocks", IssueRelationKindBlocks, []byte("blocks"), nil},
		{"depends on", IssueRelationKindDependsOn, []byte("depends on"), nil},
		{"duplicated by", IssueRelationKindDuplicatedBy, []byte("duplicated by"), nil},
		{"duplicates", IssueRelationKindDuplicates, []byte("duplicates"), nil},
		{"related to", IssueRelationKindRelatedTo, []byte("related to"), nil},
		{"subtask of", IssueRelationKindSubtaskOf, []byte("subtask of"), nil},
		{"kind high", IssueRelationKind(100), nil, ErrInvalidIssueRelationKind},
		{"kind low", IssueRelationKind(0), nil, ErrInvalidIssueRelationKind},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.r.MarshalText()
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantText, gotText)
			}
		})
	}
}

func TestIssueRelationKind_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		r       IssueRelationKind
		args    args
		wantErr error
	}{
		{"blocked by", IssueRelationKindBlockedBy, args{[]byte("blocked by")}, nil},
		{"blocks", IssueRelationKindBlocks, args{[]byte("blocks")}, nil},
		{"depends on", IssueRelationKindDependsOn, args{[]byte("depends on")}, nil},
		{"duplicated by", IssueRelationKindDuplicatedBy, args{[]byte("duplicated by")}, nil},
		{"duplicates", IssueRelationKindDuplicates, args{[]byte("duplicates")}, nil},
		{"related to", IssueRelationKindRelatedTo, args{[]byte("related to")}, nil},
		{"subtask of", IssueRelationKindSubtaskOf, args{[]byte("subtask of")}, nil},
		{"kind high", IssueRelationKind(100), args{[]byte("kind high")}, ErrInvalidIssueRelationKind},
		{"kind low", IssueRelationKind(0), args{[]byte("kind low")}, ErrInvalidIssueRelationKind},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.r.UnmarshalText(tt.args.text)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestIssueResolution_String(t *testing.T) {
	tests := []struct {
		name string
		r    IssueResolution
		want string
	}{
		{"none", IssueResolutionNone, "none"},
		{"fixed", IssueResolutionFixed, "fixed"},
		{"duplicate", IssueResolutionDuplicate, "duplicate"},
		{"won't fix", IssueResolutionWontFix, "won't fix"},
		{"invalid", IssueResolutionInvalid, "invalid"},
		{"incomplete", IssueResolutionIncomplete, "incomplete"},
		{"cannot reproduce", IssueResolutionCannotReproduce, "cannot reproduce"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.r.String())
		})
	}
}

func TestIssueResolution_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		r        IssueResolution
		wantText []byte
		wantErr  error
	}{
		{"none", IssueResolutionNone, []byte("none"), nil},
		{"fixed", IssueResolutionFixed, []byte("fixed"), nil},
		{"duplicate", IssueResolutionDuplicate, []byte("duplicate"), nil},
		{"won't fix", IssueResolutionWontFix, []byte("won't fix"), nil},
		{"invalid", IssueResolutionInvalid, []byte("invalid"), nil},
		{"incomplete", IssueResolutionIncomplete, []byte("incomplete"), nil},
		{"cannot reproduce", IssueResolutionCannotReproduce, []byte("cannot reproduce"), nil},
		{"resolution high", IssueResolution(100), []byte(""), ErrInvalidIssueResolution},
		{"resolution low", IssueResolution(0), []byte(""), ErrInvalidIssueResolution},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.r.MarshalText()
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantText, gotText)
			}
		})
	}
}

func TestIssueResolution_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		r       IssueResolution
		args    args
		wantErr error
	}{
		{"none", IssueResolutionNone, args{[]byte("none")}, nil},
		{"fixed", IssueResolutionFixed, args{[]byte("fixed")}, nil},
		{"duplicate", IssueResolutionDuplicate, args{[]byte("duplicate")}, nil},
		{"won't fix", IssueResolutionWontFix, args{[]byte("won't fix")}, nil},
		{"invalid", IssueResolutionInvalid, args{[]byte("invalid")}, nil},
		{"incomplete", IssueResolutionIncomplete, args{[]byte("incomplete")}, nil},
		{"cannot reproduce", IssueResolutionCannotReproduce, args{[]byte("cannot reproduce")}, nil},
		{"resolution high", IssueResolution(100), args{[]byte("resolution high")}, ErrInvalidIssueResolution},
		{"resolution low", IssueResolution(0), args{[]byte("resolution low")}, ErrInvalidIssueResolution},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.r.UnmarshalText(tt.args.text)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestIssueStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    IssueStatus
		want string
	}{
		{"open", IssueStatusOpen, "open"},
		{"in progress", IssueStatusInProgress, "in progress"},
		{"blocked", IssueStatusBlocked, "blocked"},
		{"review", IssueStatusReview, "review"},
		{"done", IssueStatusDone, "done"},
		{"closed", IssueStatusClosed, "closed"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestIssueStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		s        IssueStatus
		wantText []byte
		wantErr  error
	}{
		{"open", IssueStatusOpen, []byte("open"), nil},
		{"in progress", IssueStatusInProgress, []byte("in progress"), nil},
		{"blocked", IssueStatusBlocked, []byte("blocked"), nil},
		{"review", IssueStatusReview, []byte("review"), nil},
		{"done", IssueStatusDone, []byte("done"), nil},
		{"closed", IssueStatusClosed, []byte("closed"), nil},
		{"status high", IssueStatus(100), []byte(""), ErrInvalidIssueStatus},
		{"status low", IssueStatus(0), []byte(""), ErrInvalidIssueStatus},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.s.MarshalText()
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantText, gotText)
			}
		})
	}
}

func TestIssueStatus_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		s       IssueStatus
		args    args
		wantErr error
	}{
		{"open", IssueStatusOpen, args{[]byte("open")}, nil},
		{"in progress", IssueStatusInProgress, args{[]byte("in progress")}, nil},
		{"blocked", IssueStatusBlocked, args{[]byte("blocked")}, nil},
		{"review", IssueStatusReview, args{[]byte("review")}, nil},
		{"done", IssueStatusDone, args{[]byte("done")}, nil},
		{"closed", IssueStatusClosed, args{[]byte("closed")}, nil},
		{"status high", IssueStatus(100), args{[]byte("status high")}, ErrInvalidIssueStatus},
		{"status low", IssueStatus(0), args{[]byte("status low")}, ErrInvalidIssueStatus},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.s.UnmarshalText(tt.args.text)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestIssueRelation_Validate(t *testing.T) {
	tests := []struct {
		name    string
		r       IssueRelation
		wantErr error
	}{
		{
			name: "valid issue relation",
			r: IssueRelation{
				ID:     ID{inner: xid.NilID(), label: ResourceTypeIssueRelation},
				Source: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: ResourceTypeIssue},
				Target: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, label: ResourceTypeIssue},
				Kind:   IssueRelationKindRelatedTo,
			},
		},
		{
			name: "invalid issue relation kind",
			r: IssueRelation{
				ID:     ID{inner: xid.NilID(), label: ResourceTypeIssueRelation},
				Source: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: ResourceTypeIssue},
				Target: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, label: ResourceTypeIssue},
				Kind:   IssueRelationKind(100),
			},
			wantErr: ErrInvalidIssueRelationDetails,
		},
		{
			name: "invalid issue relation id",
			r: IssueRelation{
				ID:     ID{inner: xid.NilID(), label: ResourceType(0)},
				Source: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: ResourceTypeIssue},
				Target: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, label: ResourceTypeIssue},
				Kind:   IssueRelationKindRelatedTo,
			},
			wantErr: ErrInvalidIssueRelationDetails,
		},
		{
			name: "invalid issue relation source",
			r: IssueRelation{
				ID:     ID{inner: xid.NilID(), label: ResourceTypeIssueRelation},
				Source: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: ResourceType(0)},
				Target: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, label: ResourceTypeIssue},
				Kind:   IssueRelationKindRelatedTo,
			},
			wantErr: ErrInvalidIssueRelationDetails,
		},
		{
			name: "invalid issue relation target",
			r: IssueRelation{
				ID:     ID{inner: xid.NilID(), label: ResourceTypeIssueRelation},
				Source: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: ResourceTypeIssue},
				Target: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xc, 0xb, 0xa}, label: ResourceType(0)},
				Kind:   IssueRelationKindRelatedTo,
			},
			wantErr: ErrInvalidIssueRelationDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.r.Validate()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewIssueRelation(t *testing.T) {
	type args struct {
		source ID
		target ID
		kind   IssueRelationKind
	}
	tests := []struct {
		name    string
		args    args
		want    *IssueRelation
		wantErr error
	}{
		{
			name: "create new issue relation",
			args: args{
				source: ID{inner: xid.NilID(), label: ResourceTypeIssue},
				target: ID{inner: xid.NilID(), label: ResourceTypeIssue},
				kind:   IssueRelationKindRelatedTo,
			},
			want: &IssueRelation{
				ID:     ID{inner: xid.NilID(), label: ResourceTypeIssueRelation},
				Source: ID{inner: xid.NilID(), label: ResourceTypeIssue},
				Target: ID{inner: xid.NilID(), label: ResourceTypeIssue},
				Kind:   IssueRelationKindRelatedTo,
			},
		},
		{
			name: "invalid issue relation kind",
			args: args{
				source: ID{inner: xid.NilID(), label: ResourceTypeIssue},
				target: ID{inner: xid.NilID(), label: ResourceTypeIssue},
				kind:   IssueRelationKind(100),
			},
			wantErr: ErrInvalidIssueRelationDetails,
		},
		{
			name: "invalid issue relation source",
			args: args{
				source: ID{inner: xid.NilID(), label: ResourceType(0)},
				target: ID{inner: xid.NilID(), label: ResourceTypeIssue},
				kind:   IssueRelationKindRelatedTo,
			},
			wantErr: ErrInvalidIssueRelationDetails,
		},
		{
			name: "invalid issue relation target",
			args: args{
				source: ID{inner: xid.NilID(), label: ResourceTypeIssue},
				target: ID{inner: xid.NilID(), label: ResourceType(0)},
				kind:   IssueRelationKindRelatedTo,
			},
			wantErr: ErrInvalidIssueRelationDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewIssueRelation(tt.args.source, tt.args.target, tt.args.kind)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewIssue(t *testing.T) {
	type args struct {
		numericID  uint
		title      string
		kind       IssueKind
		reportedBy ID
	}
	tests := []struct {
		name    string
		args    args
		want    *Issue
		wantErr error
	}{
		{
			name: "create new issue",
			args: args{
				numericID:  1,
				title:      "title",
				kind:       IssueKindEpic,
				reportedBy: ID{inner: xid.NilID(), label: ResourceTypeUser},
			},
			want: &Issue{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
		},
		{
			name: "create new issue with invalid numeric id",
			args: args{
				numericID:  0,
				title:      "title",
				kind:       IssueKindEpic,
				reportedBy: ID{inner: xid.NilID(), label: ResourceTypeUser},
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "create new issue with invalid title",
			args: args{
				numericID:  1,
				title:      "",
				kind:       IssueKindEpic,
				reportedBy: ID{inner: xid.NilID(), label: ResourceTypeUser},
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "create new issue with invalid kind",
			args: args{
				numericID:  1,
				title:      "title",
				kind:       IssueKind(100),
				reportedBy: ID{inner: xid.NilID(), label: ResourceTypeUser},
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "create new issue with invalid reported by",
			args: args{
				numericID:  1,
				title:      "title",
				kind:       IssueKindEpic,
				reportedBy: ID{inner: xid.NilID(), label: ResourceType(0)},
			},
			wantErr: ErrInvalidIssueDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewIssue(tt.args.numericID, tt.args.title, tt.args.kind, tt.args.reportedBy)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIssue_Validate(t *testing.T) {
	type fields struct {
		ID          ID
		NumericID   uint
		Parent      *ID
		Kind        IssueKind
		Title       string
		Description string
		Status      IssueStatus
		Priority    IssuePriority
		Resolution  IssueResolution
		ReportedBy  ID
		Assignees   []ID
		Labels      []ID
		Comments    []ID
		Attachments []ID
		Watchers    []ID
		Relations   []ID
		Links       []string
		DueDate     *time.Time
		CreatedAt   *time.Time
		UpdatedAt   *time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid issue",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
		},
		{
			name: "invalid issue id",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceType(0)},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue numeric id",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   0,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue kind",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKind(100),
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue title",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue description",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "desc",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue status",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatus(100),
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue priority",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriority(100),
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue resolution",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolution(100),
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue reported by",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceType(0)},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue assignees",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees: []ID{
					{},
				},
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue labels",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels: []ID{
					{},
				},
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue comments",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments: []ID{
					{},
				},
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links:       make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue attachments",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: []ID{
					{},
				},
				Watchers:  make([]ID, 0),
				Relations: make([]ID, 0),
				Links:     make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue watchers",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers: []ID{
					{},
				},
				Relations: make([]ID, 0),
				Links:     make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue relations",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations: []ID{
					{},
				},
				Links: make([]string, 0),
			},
			wantErr: ErrInvalidIssueDetails,
		},
		{
			name: "invalid issue links",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ResourceTypeIssue},
				NumericID:   1,
				Kind:        IssueKindEpic,
				Title:       "title",
				Description: "description",
				Status:      IssueStatusOpen,
				Priority:    IssuePriorityMedium,
				Resolution:  IssueResolutionNone,
				ReportedBy:  ID{inner: xid.NilID(), label: ResourceTypeUser},
				Assignees:   make([]ID, 0),
				Labels:      make([]ID, 0),
				Comments:    make([]ID, 0),
				Attachments: make([]ID, 0),
				Watchers:    make([]ID, 0),
				Relations:   make([]ID, 0),
				Links: []string{
					"",
				},
			},
			wantErr: ErrInvalidIssueDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			i := &Issue{
				ID:          tt.fields.ID,
				NumericID:   tt.fields.NumericID,
				Parent:      tt.fields.Parent,
				Kind:        tt.fields.Kind,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				Status:      tt.fields.Status,
				Priority:    tt.fields.Priority,
				Resolution:  tt.fields.Resolution,
				ReportedBy:  tt.fields.ReportedBy,
				Assignees:   tt.fields.Assignees,
				Labels:      tt.fields.Labels,
				Comments:    tt.fields.Comments,
				Attachments: tt.fields.Attachments,
				Watchers:    tt.fields.Watchers,
				Relations:   tt.fields.Relations,
				Links:       tt.fields.Links,
				DueDate:     tt.fields.DueDate,
				CreatedAt:   tt.fields.CreatedAt,
				UpdatedAt:   tt.fields.UpdatedAt,
			}
			err := i.Validate()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
