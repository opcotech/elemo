package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    UserStatus
		want string
	}{
		{"active", UserStatusActive, "active"},
		{"pending", UserStatusPending, "pending"},
		{"inactive", UserStatusInactive, "inactive"},
		{"deleted", UserStatusDeleted, "deleted"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestUserStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       UserStatus
		want    []byte
		wantErr bool
	}{
		{"active", UserStatusActive, []byte("active"), false},
		{"pending", UserStatusPending, []byte("pending"), false},
		{"inactive", UserStatusInactive, []byte("inactive"), false},
		{"deleted", UserStatusDeleted, []byte("deleted"), false},
		{"status high", UserStatus(255), nil, true},
		{"status low", UserStatus(0), nil, true},
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

func TestUserStatus_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       *UserStatus
		text    []byte
		want    UserStatus
		wantErr bool
	}{
		{"active", new(UserStatus), []byte("active"), UserStatusActive, false},
		{"pending", new(UserStatus), []byte("pending"), UserStatusPending, false},
		{"inactive", new(UserStatus), []byte("inactive"), UserStatusInactive, false},
		{"deleted", new(UserStatus), []byte("deleted"), UserStatusDeleted, false},
		{"status high", new(UserStatus), []byte("100"), UserStatus(0), true},
		{"status low", new(UserStatus), []byte("0"), UserStatus(0), true},
		{"status invalid", new(UserStatus), []byte("invalid"), UserStatus(0), true},
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

func TestNewUser(t *testing.T) {
	type args struct {
		username string
		email    string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *User
		wantErr error
	}{
		{
			name: "create new user",
			args: args{
				username: "test",
				email:    "test@example.com",
				password: "super-secret",
			},
			want: &User{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				Links:       make([]string, 0),
				Languages:   make([]Language, 0),
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
		},
		{
			name: "create new user with empty username",
			args: args{
				username: "",
				email:    "test@example.com",
				password: "super-secret",
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "create new user with empty email",
			args: args{
				username: "test",
				email:    "",
				password: "super-secret",
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "create new user with empty password",
			args: args{
				username: "test",
				email:    "test@example.com",
				password: "",
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "create new user with invalid email",
			args: args{
				username: "test",
				email:    "test@example",
				password: "super-secret",
			},
			wantErr: ErrInvalidUserDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewUser(tt.args.username, tt.args.email, tt.args.password)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUser_Validate(t *testing.T) {
	type fields struct {
		ID          ID
		Username    string
		Email       string
		Password    string
		Status      UserStatus
		FirstName   string
		LastName    string
		Picture     string
		Title       string
		Bio         string
		Phone       string
		Address     string
		Links       []string
		Documents   []ID
		Permissions []ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "validate user",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
		},
		{
			name: "validate user with invalid ID",
			fields: fields{
				ID:          ID{},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid username",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "USERNAME",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid email",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid password",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid status",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatus(0),
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid first name",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "Johndoewhohasextremelylongnamewhichismorethanfiftycharacterslong",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid last name",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Johndoewhohasextremelylongnamewhichismorethanfiftycharacterslong",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid picture",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "example/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid title",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "T",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid bio",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid phone",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+123",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid address",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123",
				Links:       []string{"https://example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid links",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: UserIDType},
				Username:    "test",
				Email:       "test@example.com",
				Password:    "super-secret",
				Status:      UserStatusActive,
				FirstName:   "John",
				LastName:    "Doe",
				Picture:     "https://example.com/picture.jpg",
				Title:       "Software Engineer",
				Bio:         "I am a software engineer",
				Phone:       "+11234567890",
				Address:     "123 Main St, Anytown, USA",
				Links:       []string{"example.com"},
				Documents:   make([]ID, 0),
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid documents",
			fields: fields{
				ID:        ID{inner: xid.NilID(), label: UserIDType},
				Username:  "test",
				Email:     "test@example.com",
				Password:  "super-secret",
				Status:    UserStatusActive,
				FirstName: "John",
				LastName:  "Doe",
				Picture:   "https://example.com/picture.jpg",
				Title:     "Software Engineer",
				Bio:       "I am a software engineer",
				Phone:     "+11234567890",
				Address:   "123 Main St, Anytown, USA",
				Links:     []string{"https://example.com"},
				Documents: []ID{
					{},
				},
				Permissions: make([]ID, 0),
			},
			wantErr: ErrInvalidUserDetails,
		},
		{
			name: "validate user with invalid permissions",
			fields: fields{
				ID:        ID{inner: xid.NilID(), label: UserIDType},
				Username:  "test",
				Email:     "test@example.com",
				Password:  "super-secret",
				Status:    UserStatusActive,
				FirstName: "John",
				LastName:  "Doe",
				Picture:   "https://example.com/picture.jpg",
				Title:     "Software Engineer",
				Bio:       "I am a software engineer",
				Phone:     "+11234567890",
				Address:   "123 Main St, Anytown, USA",
				Links:     []string{"https://example.com"},
				Documents: make([]ID, 0),
				Permissions: []ID{
					{},
				},
			},
			wantErr: ErrInvalidUserDetails,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:          tt.fields.ID,
				Username:    tt.fields.Username,
				Email:       tt.fields.Email,
				Password:    tt.fields.Password,
				Status:      tt.fields.Status,
				FirstName:   tt.fields.FirstName,
				LastName:    tt.fields.LastName,
				Picture:     tt.fields.Picture,
				Title:       tt.fields.Title,
				Bio:         tt.fields.Bio,
				Phone:       tt.fields.Phone,
				Address:     tt.fields.Address,
				Links:       tt.fields.Links,
				Documents:   tt.fields.Documents,
				Permissions: tt.fields.Permissions,
			}
			require.ErrorIs(t, u.Validate(), tt.wantErr)
		})
	}
}
