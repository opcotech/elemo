package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFolder(t *testing.T) {
	owner := ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser}

	tests := []struct {
		name    string
		fname   string
		owner   ID
		wantErr error
	}{
		{
			name:  "create folder with valid details",
			fname: "Guides",
			owner: owner,
		},
		{
			name:    "create folder with empty name",
			fname:   "",
			owner:   owner,
			wantErr: ErrInvalidFolderDetails,
		},
		{
			name:    "create folder with nil owner",
			fname:   "Guides",
			wantErr: ErrInvalidFolderDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewFolder(tt.fname, tt.owner)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, ResourceTypeFolder, got.ID.Type)
				assert.Equal(t, tt.fname, got.Name)
				assert.Equal(t, tt.owner, got.CreatedBy)
			}
		})
	}
}

func TestFolder_Validate(t *testing.T) {
	owner := ID{Inner: xid.ID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc}, Type: ResourceTypeUser}

	tests := []struct {
		name    string
		folder  Folder
		wantErr error
	}{
		{
			name: "validate folder with valid details",
			folder: Folder{
				ID:        ID{Inner: xid.NilID(), Type: ResourceTypeFolder},
				Name:      "Guides",
				CreatedBy: owner,
			},
		},
		{
			name: "validate folder with empty name",
			folder: Folder{
				ID:        ID{Inner: xid.NilID(), Type: ResourceTypeFolder},
				Name:      "",
				CreatedBy: owner,
			},
			wantErr: ErrInvalidFolderDetails,
		},
		{
			name: "validate folder with nil owner",
			folder: Folder{
				ID:   ID{Inner: xid.NilID(), Type: ResourceTypeFolder},
				Name: "Guides",
			},
			wantErr: ErrInvalidFolderDetails,
		},
		{
			name: "validate folder with nil id",
			folder: Folder{
				ID:        ID{},
				Name:      "Guides",
				CreatedBy: owner,
			},
			wantErr: ErrInvalidFolderDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.folder.Validate()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
