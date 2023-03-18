package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func TestAssignmentKind_String(t *testing.T) {
	tests := []struct {
		name string
		k    AssignmentKind
		want string
	}{
		{"assignee", AssignmentKindAssignee, "assignee"},
		{"reviewer", AssignmentKindReviewer, "reviewer"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, tt.k.String(), "String()")
		})
	}
}

func TestAssignmentKind_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		a        AssignmentKind
		wantText []byte
		wantErr  error
	}{
		{"assignee", AssignmentKindAssignee, []byte("assignee"), nil},
		{"reviewer", AssignmentKindReviewer, []byte("reviewer"), nil},
		{"kind high", AssignmentKind(100), nil, ErrInvalidAssignmentKind},
		{"kind low", AssignmentKind(0), nil, ErrInvalidAssignmentKind},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotText, err := tt.a.MarshalText()
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantText, gotText)
			}
		})
	}
}

func TestAssignmentKind_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		a       *AssignmentKind
		text    []byte
		want    AssignmentKind
		wantErr error
	}{
		{"assignee", new(AssignmentKind), []byte("assignee"), AssignmentKindAssignee, nil},
		{"reviewer", new(AssignmentKind), []byte("reviewer"), AssignmentKindReviewer, nil},
		{"kind high", new(AssignmentKind), []byte("100"), AssignmentKind(0), ErrInvalidAssignmentKind},
		{"kind low", new(AssignmentKind), []byte("0"), AssignmentKind(0), ErrInvalidAssignmentKind},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.a.UnmarshalText(tt.text)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewAssignment(t *testing.T) {
	type args struct {
		kind     AssignmentKind
		user     ID
		resource ID
	}
	tests := []struct {
		name    string
		args    args
		want    *Assignment
		wantErr error
	}{
		{
			name: "new assignment",
			args: args{
				kind:     AssignmentKindAssignee,
				user:     ID{inner: xid.NilID(), label: ResourceTypeUser},
				resource: ID{inner: xid.NilID(), label: ResourceTypeIssue},
			},
			want: &Assignment{
				ID:       ID{inner: xid.NilID(), label: ResourceTypeAssignment},
				Kind:     AssignmentKindAssignee,
				User:     ID{inner: xid.NilID(), label: ResourceTypeUser},
				Resource: ID{inner: xid.NilID(), label: ResourceTypeIssue},
			},
		},
		{
			name: "new assignment with invalid kind",
			args: args{
				kind:     AssignmentKind(100),
				user:     ID{inner: xid.NilID(), label: ResourceTypeUser},
				resource: ID{inner: xid.NilID(), label: ResourceTypeIssue},
			},
			wantErr: ErrInvalidAssignmentDetails,
		},
		{
			name: "new assignment with invalid user id",
			args: args{
				kind:     AssignmentKindAssignee,
				user:     ID{inner: xid.NilID(), label: ResourceType(0)},
				resource: ID{inner: xid.NilID(), label: ResourceTypeIssue},
			},
			wantErr: ErrInvalidAssignmentDetails,
		},
		{
			name: "new assignment with invalid resource id",
			args: args{
				kind:     AssignmentKindAssignee,
				user:     ID{inner: xid.NilID(), label: ResourceTypeUser},
				resource: ID{inner: xid.NilID(), label: ResourceType(0)},
			},
			wantErr: ErrInvalidAssignmentDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewAssignment(tt.args.user, tt.args.resource, tt.args.kind)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAssignment_Validate(t *testing.T) {
	type fields struct {
		ID       ID
		Kind     AssignmentKind
		User     ID
		Resource ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "validate assignment with valid details",
			fields: fields{
				ID:       ID{inner: xid.NilID(), label: ResourceTypeAssignment},
				Kind:     AssignmentKindAssignee,
				User:     ID{inner: xid.NilID(), label: ResourceTypeUser},
				Resource: ID{inner: xid.NilID(), label: ResourceTypeIssue},
			},
		},
		{
			name: "validate assignment with invalid id",
			fields: fields{
				ID:       ID{inner: xid.NilID(), label: ResourceType(0)},
				Kind:     AssignmentKindAssignee,
				User:     ID{inner: xid.NilID(), label: ResourceTypeUser},
				Resource: ID{inner: xid.NilID(), label: ResourceTypeIssue},
			},
			wantErr: ErrInvalidAssignmentDetails,
		},
		{
			name: "validate assignment with invalid kind",
			fields: fields{
				ID:       ID{inner: xid.NilID(), label: ResourceTypeAssignment},
				Kind:     AssignmentKind(100),
				User:     ID{inner: xid.NilID(), label: ResourceTypeUser},
				Resource: ID{inner: xid.NilID(), label: ResourceTypeIssue},
			},
			wantErr: ErrInvalidAssignmentDetails,
		},
		{
			name: "validate assignment with invalid user id",
			fields: fields{
				ID:       ID{inner: xid.NilID(), label: ResourceTypeAssignment},
				Kind:     AssignmentKindAssignee,
				User:     ID{inner: xid.NilID(), label: ResourceType(0)},
				Resource: ID{inner: xid.NilID(), label: ResourceTypeIssue},
			},
			wantErr: ErrInvalidAssignmentDetails,
		},
		{
			name: "validate assignment with invalid resource id",
			fields: fields{
				ID:       ID{inner: xid.NilID(), label: ResourceTypeAssignment},
				Kind:     AssignmentKindAssignee,
				User:     ID{inner: xid.NilID(), label: ResourceTypeUser},
				Resource: ID{inner: xid.NilID(), label: ResourceType(0)},
			},
			wantErr: ErrInvalidAssignmentDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := &Assignment{
				ID:       tt.fields.ID,
				Kind:     tt.fields.Kind,
				User:     tt.fields.User,
				Resource: tt.fields.Resource,
			}
			err := a.Validate()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
