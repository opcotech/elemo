package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    OrganizationStatus
		want string
	}{
		{
			name: "organization status active",
			s:    OrganizationStatusActive,
			want: "active",
		},
		{
			name: "organization status deleted",
			s:    OrganizationStatusDeleted,
			want: "deleted",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestOrganizationStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       OrganizationStatus
		want    []byte
		wantErr bool
	}{
		{"active", OrganizationStatusActive, []byte("active"), false},
		{"deleted", OrganizationStatusDeleted, []byte("deleted"), false},
		{"status high", OrganizationStatus(255), nil, true},
		{"status low", OrganizationStatus(0), nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.s.MarshalText()
			if (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrganizationStatus_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       *OrganizationStatus
		text    []byte
		want    OrganizationStatus
		wantErr bool
	}{
		{"active", new(OrganizationStatus), []byte("active"), OrganizationStatusActive, false},
		{"deleted", new(OrganizationStatus), []byte("deleted"), OrganizationStatusDeleted, false},
		{"status high", new(OrganizationStatus), []byte("100"), OrganizationStatus(0), true},
		{"status low", new(OrganizationStatus), []byte("0"), OrganizationStatus(0), true},
		{"status invalid", new(OrganizationStatus), []byte("invalid"), OrganizationStatus(0), true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.s.UnmarshalText(tt.text); (err != nil) != tt.wantErr {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewOrganization(t *testing.T) {
	type args struct {
		name  string
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    *Organization
		wantErr error
	}{
		{
			name: "create organization",
			args: args{
				name:  "test",
				email: "info@example.com",
			},
			want: &Organization{
				ID:         ID{inner: xid.NilID(), label: ResourceTypeOrganization},
				Name:       "test",
				Email:      "info@example.com",
				Status:     OrganizationStatusActive,
				Namespaces: make([]ID, 0),
				Members:    make([]ID, 0),
				Teams:      make([]ID, 0),
			},
		},
		{
			name: "create organization with empty name",
			args: args{
				name:  "",
				email: "info@example.com",
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
		{
			name: "create organization with empty email",
			args: args{
				name:  "test",
				email: "",
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
		{
			name: "create organization with invalid email",
			args: args{
				name:  "test",
				email: "invalid@example",
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewOrganization(tt.args.name, tt.args.email)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestOrganization_Validate(t *testing.T) {
	type fields struct {
		ID          ID
		Name        string
		Email       string
		Description string
		Status      OrganizationStatus
		Namespaces  []ID
		Members     []ID
		Teams       []ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid organization",
			fields: fields{
				ID:         ID{inner: xid.NilID(), label: ResourceTypeOrganization},
				Name:       "test",
				Email:      "test@example.com",
				Status:     OrganizationStatusActive,
				Namespaces: make([]ID, 0),
				Members:    make([]ID, 0),
				Teams:      make([]ID, 0),
			},
		},
		{
			name: "invalid organization id",
			fields: fields{
				ID:         ID{inner: xid.NilID(), label: ResourceType(0)},
				Name:       "test",
				Email:      "test@example.com",
				Status:     OrganizationStatusActive,
				Namespaces: make([]ID, 0),
				Members:    make([]ID, 0),
				Teams:      make([]ID, 0),
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
		{
			name: "invalid organization email",
			fields: fields{
				ID:         ID{inner: xid.NilID(), label: ResourceType(0)},
				Name:       "test",
				Email:      "test.com",
				Status:     OrganizationStatusActive,
				Namespaces: make([]ID, 0),
				Members:    make([]ID, 0),
				Teams:      make([]ID, 0),
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
		{
			name: "invalid organization status",
			fields: fields{
				ID:         ID{inner: xid.NilID(), label: ResourceType(0)},
				Name:       "test",
				Email:      "test@example.com",
				Status:     OrganizationStatus(0),
				Namespaces: make([]ID, 0),
				Members:    make([]ID, 0),
				Teams:      make([]ID, 0),
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
		{
			name: "invalid namespaces",
			fields: fields{
				ID:     ID{inner: xid.NilID(), label: ResourceTypeOrganization},
				Name:   "test",
				Email:  "test@example.com",
				Status: OrganizationStatusActive,
				Namespaces: []ID{
					{},
				},
				Members: make([]ID, 0),
				Teams:   make([]ID, 0),
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
		{
			name: "invalid members",
			fields: fields{
				ID:         ID{inner: xid.NilID(), label: ResourceTypeOrganization},
				Name:       "test",
				Email:      "test@example.com",
				Status:     OrganizationStatusActive,
				Namespaces: make([]ID, 0),
				Members: []ID{
					{},
				},
				Teams: make([]ID, 0),
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
		{
			name: "invalid teams",
			fields: fields{
				ID:         ID{inner: xid.NilID(), label: ResourceTypeOrganization},
				Name:       "test",
				Email:      "test@example.com",
				Status:     OrganizationStatusActive,
				Namespaces: make([]ID, 0),
				Members:    make([]ID, 0),
				Teams: []ID{
					{},
				},
			},
			wantErr: ErrInvalidOrganizationDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := &Organization{
				ID:         tt.fields.ID,
				Name:       tt.fields.Name,
				Email:      tt.fields.Email,
				Status:     tt.fields.Status,
				Namespaces: tt.fields.Namespaces,
				Members:    tt.fields.Members,
				Teams:      tt.fields.Teams,
			}
			require.ErrorIs(t, o.Validate(), tt.wantErr)
		})
	}
}
