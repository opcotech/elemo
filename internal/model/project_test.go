package model

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    ProjectStatus
		want string
	}{
		{"active", ProjectStatusActive, "active"},
		{"pending", ProjectStatusPending, "pending"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.s.String())
		})
	}
}

func TestProjectStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       ProjectStatus
		want    []byte
		wantErr bool
	}{
		{"active", ProjectStatusActive, []byte("active"), false},
		{"pending", ProjectStatusPending, []byte("pending"), false},
		{"status high", ProjectStatus(255), nil, true},
		{"status low", ProjectStatus(0), nil, true},
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

func TestProjectStatus_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		s       *ProjectStatus
		text    []byte
		want    ProjectStatus
		wantErr bool
	}{
		{"active", new(ProjectStatus), []byte("active"), ProjectStatusActive, false},
		{"pending", new(ProjectStatus), []byte("pending"), ProjectStatusPending, false},
		{"status high", new(ProjectStatus), []byte("100"), ProjectStatus(0), true},
		{"status low", new(ProjectStatus), []byte("0"), ProjectStatus(0), true},
		{"status invalid", new(ProjectStatus), []byte("invalid"), ProjectStatus(0), true},
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

func TestNewProject(t *testing.T) {
	type args struct {
		key  string
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *Project
		wantErr error
	}{
		{
			name: "create new project",
			args: args{
				key:  "test",
				name: "Test Project",
			},
			want: &Project{
				ID:        ID{inner: xid.NilID(), label: ProjectIDType},
				Key:       "test",
				Name:      "Test Project",
				Status:    ProjectStatusActive,
				Teams:     make([]ID, 0),
				Documents: make([]ID, 0),
				Issues:    make([]ID, 0),
			},
		},
		{
			name: "create new project with invalid key",
			args: args{
				key:  "test!",
				name: "Test Project",
			},
			wantErr: ErrInvalidProjectDetails,
		},
		{
			name: "create new project with invalid name",
			args: args{
				key:  "test",
				name: "",
			},
			wantErr: ErrInvalidProjectDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewProject(tt.args.key, tt.args.name)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestProject_Validate(t *testing.T) {
	type fields struct {
		ID          ID
		Key         string
		Name        string
		Description string
		Status      ProjectStatus
		Teams       []ID
		Documents   []ID
		Issues      []ID
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid project",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: OrganizationIDType},
				Key:         "test",
				Name:        "test",
				Description: "Test description",
				Status:      ProjectStatusActive,
				Teams:       make([]ID, 0),
				Documents:   make([]ID, 0),
				Issues:      make([]ID, 0),
			},
		},
		{
			name: "invalid project id",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ""},
				Key:         "test",
				Name:        "test",
				Description: "Test description",
				Status:      ProjectStatusActive,
				Teams:       make([]ID, 0),
				Documents:   make([]ID, 0),
				Issues:      make([]ID, 0),
			},
			wantErr: ErrInvalidProjectDetails,
		},
		{
			name: "invalid project key",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ""},
				Key:         "tst",
				Name:        "test",
				Description: "Test description",
				Status:      ProjectStatusActive,
				Teams:       make([]ID, 0),
				Documents:   make([]ID, 0),
				Issues:      make([]ID, 0),
			},
			wantErr: ErrInvalidProjectDetails,
		},
		{
			name: "invalid project name",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ""},
				Key:         "test",
				Name:        "t",
				Description: "Test description",
				Status:      ProjectStatusActive,
				Teams:       make([]ID, 0),
				Documents:   make([]ID, 0),
				Issues:      make([]ID, 0),
			},
			wantErr: ErrInvalidProjectDetails,
		},
		{
			name: "invalid project description",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ""},
				Key:         "test",
				Name:        "test",
				Description: "Test",
				Status:      ProjectStatusActive,
				Teams:       make([]ID, 0),
				Documents:   make([]ID, 0),
				Issues:      make([]ID, 0),
			},
			wantErr: ErrInvalidProjectDetails,
		},
		{
			name: "invalid project status",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: ""},
				Key:         "test",
				Name:        "test",
				Description: "Test description",
				Status:      ProjectStatus(0),
				Teams:       make([]ID, 0),
				Documents:   make([]ID, 0),
				Issues:      make([]ID, 0),
			},
			wantErr: ErrInvalidProjectDetails,
		},
		{
			name: "invalid project teams",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: OrganizationIDType},
				Key:         "test",
				Name:        "test",
				Description: "Test description",
				Status:      ProjectStatusActive,
				Teams: []ID{
					{},
				},
				Documents: make([]ID, 0),
				Issues:    make([]ID, 0),
			},
			wantErr: ErrInvalidProjectDetails,
		},
		{
			name: "invalid project documents",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: OrganizationIDType},
				Key:         "test",
				Name:        "test",
				Description: "Test description",
				Status:      ProjectStatusActive,
				Teams:       make([]ID, 0),
				Documents: []ID{
					{},
				},
				Issues: make([]ID, 0),
			},
			wantErr: ErrInvalidProjectDetails,
		},
		{
			name: "invalid project issues",
			fields: fields{
				ID:          ID{inner: xid.NilID(), label: OrganizationIDType},
				Key:         "test",
				Name:        "test",
				Description: "Test description",
				Status:      ProjectStatusActive,
				Teams:       make([]ID, 0),
				Documents:   make([]ID, 0),
				Issues: []ID{
					{},
				},
			},
			wantErr: ErrInvalidProjectDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &Project{
				ID:          tt.fields.ID,
				Key:         tt.fields.Key,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Status:      tt.fields.Status,
				Teams:       tt.fields.Teams,
				Documents:   tt.fields.Documents,
				Issues:      tt.fields.Issues,
			}
			require.ErrorIs(t, p.Validate(), tt.wantErr)
		})
	}
}
