package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAttachment(t *testing.T) {
	type args struct {
		name   string
		fileID string
		owner  ID
	}
	tests := []struct {
		name    string
		args    args
		want    *Attachment
		wantErr error
	}{
		{
			name: "create attachment with valid details",
			args: args{
				name:   "test",
				fileID: "file_id",
				owner:  ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: "User"},
			},
			want: &Attachment{
				ID:        ID{inner: xid.NilID(), label: AttachmentIDType},
				Name:      "test",
				FileID:    "file_id",
				CreatedBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: "User"},
			},
		},
		{
			name: "create attachment with invalid name",
			args: args{
				name:   "t",
				fileID: "file_id",
			},
			wantErr: ErrInvalidAttachmentDetails,
		},
		{
			name: "create attachment with empty name",
			args: args{
				name:   "",
				fileID: "file_id",
			},
			wantErr: ErrInvalidAttachmentDetails,
		},
		{
			name: "create attachment with nil owner",
			args: args{
				name:   "test",
				fileID: "file_id",
			},
			wantErr: ErrInvalidAttachmentDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewAttachment(tt.args.name, tt.args.fileID, tt.args.owner)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAttachment_Validate(t *testing.T) {
	type fields struct {
		ID      ID
		Name    string
		FileID  string
		OwnedBy ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "validate attachment with valid details",
			fields: fields{
				ID:      ID{inner: xid.NilID(), label: AttachmentIDType},
				Name:    "test",
				FileID:  "file_id",
				OwnedBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: "User"},
			},
		},
		{
			name: "validate attachment with invalid name",
			fields: fields{
				ID:      ID{inner: xid.NilID(), label: AttachmentIDType},
				Name:    "t",
				FileID:  "file_id",
				OwnedBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: "User"},
			},
			wantErr: ErrInvalidAttachmentDetails,
		},
		{
			name: "validate attachment with empty name",
			fields: fields{
				ID:      ID{inner: xid.NilID(), label: AttachmentIDType},
				Name:    "",
				FileID:  "file_id",
				OwnedBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: "User"},
			},
			wantErr: ErrInvalidAttachmentDetails,
		},
		{
			name: "validate attachment with nil owner",
			fields: fields{
				ID:      ID{inner: xid.NilID(), label: AttachmentIDType},
				Name:    "test",
				FileID:  "file_id",
				OwnedBy: ID{},
			},
			wantErr: ErrInvalidAttachmentDetails,
		},
		{
			name: "validate attachment with nil file id",
			fields: fields{
				ID:      ID{inner: xid.NilID(), label: AttachmentIDType},
				Name:    "test",
				FileID:  "",
				OwnedBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: "User"},
			},
			wantErr: ErrInvalidAttachmentDetails,
		},
		{
			name: "validate attachment with nil id",
			fields: fields{
				ID:      ID{},
				Name:    "test",
				FileID:  "file_id",
				OwnedBy: ID{inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, label: "User"},
			},
			wantErr: ErrInvalidAttachmentDetails,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Attachment{
				ID:        tt.fields.ID,
				Name:      tt.fields.Name,
				FileID:    tt.fields.FileID,
				CreatedBy: tt.fields.OwnedBy,
			}
			err := d.Validate()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
