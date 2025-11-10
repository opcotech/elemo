package model

import (
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/pkg/convert"
)

func TestNewNamespace(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *Namespace
		wantErr error
	}{
		{
			name: "create namespace with valid details",
			args: args{
				name: "test",
			},
			want: &Namespace{
				ID:        ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:      "test",
				Projects:  make([]*NamespaceProject, 0),
				Documents: make([]*NamespaceDocument, 0),
			},
		},
		{
			name: "create namespace with invalid name",
			args: args{
				name: "t",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
		{
			name: "create namespace with empty name",
			args: args{
				name: "",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewNamespace(tt.args.name)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.want != nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNamespace_Validate(t *testing.T) {
	type fields struct {
		ID          ID
		Name        string
		Description string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "validate namespace with valid details",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:        "test",
				Description: "test description",
			},
		},
		{
			name: "validate namespace with invalid ID",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceType(0)},
				Name:        "test",
				Description: "test description",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
		{
			name: "validate namespace with invalid name",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:        "t",
				Description: "test description",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
		{
			name: "validate namespace with empty name",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:        "",
				Description: "test description",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
		{
			name: "validate namespace with invalid description",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceTypeNamespace},
				Name:        "test",
				Description: "t",
			},
			wantErr: ErrInvalidNamespaceDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := &Namespace{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
			}
			require.ErrorIs(t, n.Validate(), tt.wantErr)
		})
	}
}

func TestNamespaceProject_Validate(t *testing.T) {
	type fields struct {
		ID          ID
		Key         string
		Name        string
		Description string
		Logo        string
		Status      ProjectStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "validate namespace project with valid details",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
		},
		{
			name: "validate namespace project with minimal valid details",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "",
				Logo:        "",
				Status:      ProjectStatusActive,
			},
		},
		{
			name: "validate namespace project with invalid ID type",
			fields: fields{
				ID:          MustNewID(ResourceTypeNamespace),
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid ID",
			fields: fields{
				ID:          ID{Inner: xid.NilID(), Type: ResourceType(0)},
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid key (too short)",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "EN",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid key (too long)",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "ENGINEER",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid key (non-alpha)",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "EN1",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid name (too short)",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "ENG",
				Name:        "EN",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid name (too long)",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "ENG",
				Name:        string(make([]byte, 121)),
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid description (too short)",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "short",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid logo URL",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "not-a-url",
				Status:      ProjectStatusActive,
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "validate namespace project with invalid status",
			fields: fields{
				ID:          MustNewID(ResourceTypeProject),
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatus(0),
			},
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			np := &NamespaceProject{
				ID:          tt.fields.ID,
				Key:         tt.fields.Key,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Logo:        tt.fields.Logo,
				Status:      tt.fields.Status,
			}
			err := np.Validate()
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewNamespaceProject(t *testing.T) {
	type args struct {
		id          ID
		key         string
		name        string
		description string
		logo        string
		status      ProjectStatus
	}
	tests := []struct {
		name    string
		args    args
		want    *NamespaceProject
		wantErr error
	}{
		{
			name: "create new namespace project",
			args: args{
				id:          MustNewID(ResourceTypeProject),
				key:         "ENG",
				name:        "Engineering Project",
				description: "Engineering team project",
				logo:        "https://example.com/logo.png",
				status:      ProjectStatusActive,
			},
			want: &NamespaceProject{
				ID:          ID{},
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusActive,
			},
			wantErr: nil,
		},
		{
			name: "create new namespace project with minimal fields",
			args: args{
				id:          MustNewID(ResourceTypeProject),
				key:         "ENG",
				name:        "Engineering Project",
				description: "",
				logo:        "",
				status:      ProjectStatusActive,
			},
			want: &NamespaceProject{
				ID:          ID{},
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "",
				Logo:        "",
				Status:      ProjectStatusActive,
			},
			wantErr: nil,
		},
		{
			name: "create new namespace project with pending status",
			args: args{
				id:          MustNewID(ResourceTypeProject),
				key:         "ENG",
				name:        "Engineering Project",
				description: "Engineering team project",
				logo:        "https://example.com/logo.png",
				status:      ProjectStatusPending,
			},
			want: &NamespaceProject{
				ID:          ID{},
				Key:         "ENG",
				Name:        "Engineering Project",
				Description: "Engineering team project",
				Logo:        "https://example.com/logo.png",
				Status:      ProjectStatusPending,
			},
			wantErr: nil,
		},
		{
			name: "create new namespace project with invalid ID type",
			args: args{
				id:          MustNewID(ResourceTypeNamespace),
				key:         "ENG",
				name:        "Engineering Project",
				description: "Engineering team project",
				logo:        "https://example.com/logo.png",
				status:      ProjectStatusActive,
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "create new namespace project with invalid key",
			args: args{
				id:          MustNewID(ResourceTypeProject),
				key:         "EN",
				name:        "Engineering Project",
				description: "Engineering team project",
				logo:        "https://example.com/logo.png",
				status:      ProjectStatusActive,
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "create new namespace project with invalid name",
			args: args{
				id:          MustNewID(ResourceTypeProject),
				key:         "ENG",
				name:        "EN",
				description: "Engineering team project",
				logo:        "https://example.com/logo.png",
				status:      ProjectStatusActive,
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "create new namespace project with invalid description",
			args: args{
				id:          MustNewID(ResourceTypeProject),
				key:         "ENG",
				name:        "Engineering Project",
				description: "short",
				logo:        "https://example.com/logo.png",
				status:      ProjectStatusActive,
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "create new namespace project with invalid logo",
			args: args{
				id:          MustNewID(ResourceTypeProject),
				key:         "ENG",
				name:        "Engineering Project",
				description: "Engineering team project",
				logo:        "not-a-url",
				status:      ProjectStatusActive,
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
		{
			name: "create new namespace project with invalid status",
			args: args{
				id:          MustNewID(ResourceTypeProject),
				key:         "ENG",
				name:        "Engineering Project",
				description: "Engineering team project",
				logo:        "https://example.com/logo.png",
				status:      ProjectStatus(0),
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceProjectDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewNamespaceProject(tt.args.id, tt.args.key, tt.args.name, tt.args.description, tt.args.logo, tt.args.status)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.want != nil {
				assert.Equal(t, tt.want.Key, got.Key)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Description, got.Description)
				assert.Equal(t, tt.want.Logo, got.Logo)
				assert.Equal(t, tt.want.Status, got.Status)
				// ID will be set from the argument
				assert.Equal(t, tt.args.id, got.ID)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestNamespaceDocument_Validate(t *testing.T) {
	type fields struct {
		ID        ID
		Name      string
		Excerpt   string
		CreatedBy ID
		CreatedAt *time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "validate namespace document with valid details",
			fields: fields{
				ID:        MustNewID(ResourceTypeDocument),
				Name:      "Project Plan",
				Excerpt:   "Overview of the project plan and goals",
				CreatedBy: MustNewID(ResourceTypeUser),
				CreatedAt: convert.ToPointer(time.Now()),
			},
		},
		{
			name: "validate namespace document with minimal valid details",
			fields: fields{
				ID:        MustNewID(ResourceTypeDocument),
				Name:      "Project Plan",
				Excerpt:   "",
				CreatedBy: MustNewID(ResourceTypeUser),
				CreatedAt: nil,
			},
		},
		{
			name: "validate namespace document with invalid ID type",
			fields: fields{
				ID:        MustNewID(ResourceTypeNamespace),
				Name:      "Project Plan",
				Excerpt:   "Overview of the project plan and goals",
				CreatedBy: MustNewID(ResourceTypeUser),
				CreatedAt: convert.ToPointer(time.Now()),
			},
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
		{
			name: "validate namespace document with invalid ID",
			fields: fields{
				ID:        ID{Inner: xid.NilID(), Type: ResourceType(0)},
				Name:      "Project Plan",
				Excerpt:   "Overview of the project plan and goals",
				CreatedBy: MustNewID(ResourceTypeUser),
				CreatedAt: convert.ToPointer(time.Now()),
			},
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
		{
			name: "validate namespace document with invalid name (too short)",
			fields: fields{
				ID:        MustNewID(ResourceTypeDocument),
				Name:      "AB",
				Excerpt:   "Overview of the project plan and goals",
				CreatedBy: MustNewID(ResourceTypeUser),
				CreatedAt: convert.ToPointer(time.Now()),
			},
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
		{
			name: "validate namespace document with invalid name (too long)",
			fields: fields{
				ID:        MustNewID(ResourceTypeDocument),
				Name:      string(make([]byte, 121)),
				Excerpt:   "Overview of the project plan and goals",
				CreatedBy: MustNewID(ResourceTypeUser),
				CreatedAt: convert.ToPointer(time.Now()),
			},
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
		{
			name: "validate namespace document with invalid excerpt (too short)",
			fields: fields{
				ID:        MustNewID(ResourceTypeDocument),
				Name:      "Project Plan",
				Excerpt:   "short",
				CreatedBy: MustNewID(ResourceTypeUser),
				CreatedAt: convert.ToPointer(time.Now()),
			},
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
		{
			name: "validate namespace document with invalid created_by ID",
			fields: fields{
				ID:        MustNewID(ResourceTypeDocument),
				Name:      "Project Plan",
				Excerpt:   "Overview of the project plan and goals",
				CreatedBy: ID{Inner: xid.NilID(), Type: ResourceType(0)},
				CreatedAt: convert.ToPointer(time.Now()),
			},
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			nd := &NamespaceDocument{
				ID:        tt.fields.ID,
				Name:      tt.fields.Name,
				Excerpt:   tt.fields.Excerpt,
				CreatedBy: tt.fields.CreatedBy,
				CreatedAt: tt.fields.CreatedAt,
			}
			err := nd.Validate()
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewNamespaceDocument(t *testing.T) {
	type args struct {
		id        ID
		name      string
		excerpt   string
		createdBy ID
		createdAt *time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    *NamespaceDocument
		wantErr error
	}{
		{
			name: "create new namespace document",
			args: args{
				id:        MustNewID(ResourceTypeDocument),
				name:      "Project Plan",
				excerpt:   "Overview of the project plan and goals",
				createdBy: MustNewID(ResourceTypeUser),
				createdAt: convert.ToPointer(time.Now()),
			},
			want: &NamespaceDocument{
				ID:        ID{},
				Name:      "Project Plan",
				Excerpt:   "Overview of the project plan and goals",
				CreatedBy: ID{},
				CreatedAt: convert.ToPointer(time.Now()),
			},
			wantErr: nil,
		},
		{
			name: "create new namespace document with minimal fields",
			args: args{
				id:        MustNewID(ResourceTypeDocument),
				name:      "Project Plan",
				excerpt:   "",
				createdBy: MustNewID(ResourceTypeUser),
				createdAt: nil,
			},
			want: &NamespaceDocument{
				ID:        ID{},
				Name:      "Project Plan",
				Excerpt:   "",
				CreatedBy: ID{},
				CreatedAt: nil,
			},
			wantErr: nil,
		},
		{
			name: "create new namespace document with invalid ID type",
			args: args{
				id:        MustNewID(ResourceTypeNamespace),
				name:      "Project Plan",
				excerpt:   "Overview of the project plan and goals",
				createdBy: MustNewID(ResourceTypeUser),
				createdAt: convert.ToPointer(time.Now()),
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
		{
			name: "create new namespace document with invalid name",
			args: args{
				id:        MustNewID(ResourceTypeDocument),
				name:      "AB",
				excerpt:   "Overview of the project plan and goals",
				createdBy: MustNewID(ResourceTypeUser),
				createdAt: convert.ToPointer(time.Now()),
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
		{
			name: "create new namespace document with invalid excerpt",
			args: args{
				id:        MustNewID(ResourceTypeDocument),
				name:      "Project Plan",
				excerpt:   "short",
				createdBy: MustNewID(ResourceTypeUser),
				createdAt: convert.ToPointer(time.Now()),
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
		{
			name: "create new namespace document with invalid created_by ID",
			args: args{
				id:        MustNewID(ResourceTypeDocument),
				name:      "Project Plan",
				excerpt:   "Overview of the project plan and goals",
				createdBy: ID{Inner: xid.NilID(), Type: ResourceType(0)},
				createdAt: convert.ToPointer(time.Now()),
			},
			want:    nil,
			wantErr: ErrInvalidNamespaceDocumentDetails,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewNamespaceDocument(tt.args.id, tt.args.name, tt.args.excerpt, tt.args.createdBy, tt.args.createdAt)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.want != nil {
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Excerpt, got.Excerpt)
				// ID and CreatedBy will be set from the arguments
				assert.Equal(t, tt.args.id, got.ID)
				assert.Equal(t, tt.args.createdBy, got.CreatedBy)
				// CreatedAt comparison
				if tt.want.CreatedAt != nil {
					assert.NotNil(t, got.CreatedAt)
				} else {
					assert.Nil(t, got.CreatedAt)
				}
			} else {
				assert.Nil(t, got)
			}
		})
	}
}
