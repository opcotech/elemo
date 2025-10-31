package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/pkg/convert"
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
				ID:         ID{Inner: xid.NilID(), Type: ResourceTypeOrganization},
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
				ID:         ID{Inner: xid.NilID(), Type: ResourceTypeOrganization},
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
				ID:         ID{Inner: xid.NilID(), Type: ResourceType(0)},
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
				ID:         ID{Inner: xid.NilID(), Type: ResourceType(0)},
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
				ID:         ID{Inner: xid.NilID(), Type: ResourceType(0)},
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
				ID:     ID{Inner: xid.NilID(), Type: ResourceTypeOrganization},
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
				ID:         ID{Inner: xid.NilID(), Type: ResourceTypeOrganization},
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
				ID:         ID{Inner: xid.NilID(), Type: ResourceTypeOrganization},
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

func TestOrganizationMember_Validate(t *testing.T) {
	type fields struct {
		ID        ID
		FirstName string
		LastName  string
		Email     string
		Picture   *string
		Status    UserStatus
		Roles     []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid organization member",
			fields: fields{
				ID:        MustNewID(ResourceTypeUser),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Picture:   nil,
				Status:    UserStatusActive,
				Roles:     []string{"owner", "admin"},
			},
			wantErr: nil,
		},
		{
			name: "valid organization member with picture",
			fields: fields{
				ID:        MustNewID(ResourceTypeUser),
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane.smith@example.com",
				Picture:   convert.ToPointer("https://example.com/picture.jpg"),
				Status:    UserStatusActive,
				Roles:     []string{"member"},
			},
			wantErr: nil,
		},
		{
			name: "invalid ID (wrong resource type)",
			fields: fields{
				ID:        MustNewID(ResourceTypeOrganization),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Picture:   nil,
				Status:    UserStatusActive,
				Roles:     []string{"owner"},
			},
			wantErr: ErrInvalidOrganizationMemberDetails,
		},
		{
			name: "invalid first name (empty)",
			fields: fields{
				ID:        MustNewID(ResourceTypeUser),
				FirstName: "",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Picture:   nil,
				Status:    UserStatusActive,
				Roles:     []string{"owner"},
			},
			wantErr: ErrInvalidOrganizationMemberDetails,
		},
		{
			name: "invalid last name (empty)",
			fields: fields{
				ID:        MustNewID(ResourceTypeUser),
				FirstName: "John",
				LastName:  "",
				Email:     "john.doe@example.com",
				Picture:   nil,
				Status:    UserStatusActive,
				Roles:     []string{"owner"},
			},
			wantErr: ErrInvalidOrganizationMemberDetails,
		},
		{
			name: "invalid email",
			fields: fields{
				ID:        MustNewID(ResourceTypeUser),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "invalid-email",
				Picture:   nil,
				Status:    UserStatusActive,
				Roles:     []string{"owner"},
			},
			wantErr: ErrInvalidOrganizationMemberDetails,
		},
		{
			name: "invalid picture URL",
			fields: fields{
				ID:        MustNewID(ResourceTypeUser),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Picture:   convert.ToPointer("not-a-url"),
				Status:    UserStatusActive,
				Roles:     []string{"owner"},
			},
			wantErr: ErrInvalidOrganizationMemberDetails,
		},
		{
			name: "invalid status",
			fields: fields{
				ID:        MustNewID(ResourceTypeUser),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Picture:   nil,
				Status:    UserStatus(0),
				Roles:     []string{"owner"},
			},
			wantErr: ErrInvalidOrganizationMemberDetails,
		},
		{
			name: "empty roles",
			fields: fields{
				ID:        MustNewID(ResourceTypeUser),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Picture:   nil,
				Status:    UserStatusActive,
				Roles:     []string{},
			},
			wantErr: nil, // Empty roles are allowed
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			om := &OrganizationMember{
				ID:        tt.fields.ID,
				FirstName: tt.fields.FirstName,
				LastName:  tt.fields.LastName,
				Email:     tt.fields.Email,
				Picture:   tt.fields.Picture,
				Status:    tt.fields.Status,
				Roles:     tt.fields.Roles,
			}
			err := om.Validate()
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewOrganizationMember(t *testing.T) {
	type args struct {
		id        ID
		firstName string
		lastName  string
		email     string
		picture   *string
		status    UserStatus
		roles     []string
	}
	tests := []struct {
		name    string
		args    args
		want    *OrganizationMember
		wantErr error
	}{
		{
			name: "create new organization member",
			args: args{
				id:        MustNewID(ResourceTypeUser),
				firstName: "John",
				lastName:  "Doe",
				email:     "john.doe@example.com",
				picture:   nil,
				status:    UserStatusActive,
				roles:     []string{"owner"},
			},
			want: &OrganizationMember{
				ID:        ID{},
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Picture:   nil,
				Status:    UserStatusActive,
				Roles:     []string{"owner"},
			},
			wantErr: nil,
		},
		{
			name: "create new organization member with picture",
			args: args{
				id:        MustNewID(ResourceTypeUser),
				firstName: "Jane",
				lastName:  "Smith",
				email:     "jane.smith@example.com",
				picture:   convert.ToPointer("https://example.com/picture.jpg"),
				status:    UserStatusActive,
				roles:     []string{"admin", "member"},
			},
			want: &OrganizationMember{
				ID:        ID{},
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane.smith@example.com",
				Picture:   convert.ToPointer("https://example.com/picture.jpg"),
				Status:    UserStatusActive,
				Roles:     []string{"admin", "member"},
			},
			wantErr: nil,
		},
		{
			name: "create new organization member with invalid ID",
			args: args{
				id:        MustNewID(ResourceTypeOrganization),
				firstName: "John",
				lastName:  "Doe",
				email:     "john.doe@example.com",
				picture:   nil,
				status:    UserStatusActive,
				roles:     []string{"owner"},
			},
			want:    nil,
			wantErr: ErrInvalidOrganizationMemberDetails,
		},
		{
			name: "create new organization member with invalid email",
			args: args{
				id:        MustNewID(ResourceTypeUser),
				firstName: "John",
				lastName:  "Doe",
				email:     "invalid-email",
				picture:   nil,
				status:    UserStatusActive,
				roles:     []string{"owner"},
			},
			want:    nil,
			wantErr: ErrInvalidOrganizationMemberDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewOrganizationMember(
				tt.args.id,
				tt.args.firstName,
				tt.args.lastName,
				tt.args.email,
				tt.args.picture,
				tt.args.status,
				tt.args.roles,
			)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				// Use the ID from args instead of want, since want.ID was generated separately
				assert.Equal(t, tt.args.id, got.ID)
				assert.Equal(t, tt.want.FirstName, got.FirstName)
				assert.Equal(t, tt.want.LastName, got.LastName)
				assert.Equal(t, tt.want.Email, got.Email)
				if tt.want.Picture == nil {
					assert.Nil(t, got.Picture)
				} else {
					assert.Equal(t, *tt.want.Picture, *got.Picture)
				}
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.Roles, got.Roles)
			}
		})
	}
}
