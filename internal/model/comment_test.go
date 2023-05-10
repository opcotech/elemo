package model

import (
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComment(t *testing.T) {
	type args struct {
		content   string
		createdBy ID
	}
	tests := []struct {
		name    string
		args    args
		want    *Comment
		wantErr error
	}{
		{
			name: "create comment with valid details",
			args: args{
				content:   "testing",
				createdBy: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
			},
			want: &Comment{
				ID:        ID{Inner: xid.NilID(), Type: ResourceTypeComment},
				Content:   "testing",
				CreatedBy: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
			},
		},
		{
			name: "create comment with invalid creator",
			args: args{
				content:   "testing",
				createdBy: ID{},
			},
			wantErr: ErrInvalidCommentDetails,
		},
		{
			name: "create comment with invalid content",
			args: args{
				content: "t",
			},
			wantErr: ErrInvalidCommentDetails,
		},
		{
			name: "create comment with empty content",
			args: args{
				content: "",
			},
			wantErr: ErrInvalidCommentDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewComment(tt.args.content, tt.args.createdBy)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestComment_Validate(t *testing.T) {
	type fields struct {
		ID        ID
		Content   string
		CreatedBy ID
		CreatedAt *time.Time
		UpdatedAt *time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "validate comment with valid details",
			fields: fields{
				ID:        ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeComment},
				Content:   "testing",
				CreatedBy: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
			},
		},
		{
			name: "validate comment with invalid ID",
			fields: fields{
				ID:        ID{},
				Content:   "testing",
				CreatedBy: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
			},
			wantErr: ErrInvalidCommentDetails,
		},
		{
			name: "validate comment with invalid content",
			fields: fields{
				ID:        ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeComment},
				Content:   "t",
				CreatedBy: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
			},
			wantErr: ErrInvalidCommentDetails,
		},
		{
			name: "validate comment with empty content",
			fields: fields{
				ID:        ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeComment},
				Content:   "",
				CreatedBy: ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser},
			},
			wantErr: ErrInvalidCommentDetails,
		},
		{
			name: "validate comment with invalid creator",
			fields: fields{
				ID:        ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeComment},
				Content:   "testing",
				CreatedBy: ID{},
			},
			wantErr: ErrInvalidCommentDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Comment{
				ID:        tt.fields.ID,
				Content:   tt.fields.Content,
				CreatedBy: tt.fields.CreatedBy,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
			}
			err := c.Validate()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
